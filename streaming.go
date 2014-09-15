package oanda

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

const (
	DefaultStallTimeout = 10 * time.Second
)

type HeartbeatHandleFunc func(t time.Time)
type messageHandleFunc func(key string, msgData json.RawMessage) error

type streamServer struct {
	StallTimeout  time.Duration
	HeartbeatFunc HeartbeatHandleFunc

	ctx        *Context
	hbCh       chan time.Time
	isStopped  bool
	stallTimer *time.Timer
	wg         sync.WaitGroup
	rspMtx     sync.Mutex
	rsp        *http.Response
}

func (c *Client) newStreamServer(u *url.URL) (*streamServer, error) {
	ctx, err := c.newContext("GET", u, nil)
	if err != nil {
		return nil, err
	}
	ss := streamServer{
		StallTimeout: DefaultStallTimeout,
		ctx:          ctx,
		hbCh:         make(chan time.Time),
		isStopped:    true,
	}
	return &ss, nil
}

func (ss *streamServer) Run(handleFn messageHandleFunc) error {
	err := ss.init()
	if err != nil {
		return err
	}
	defer ss.cleanup()

	for !ss.isStopped {
		ss.connect()
		err = ss.runDispatchLoop(handleFn)
		if err != nil {
			// FIXME: Log message
		}
	}
	return err
}

func (ss *streamServer) runDispatchLoop(handleFn messageHandleFunc) error {
	var strm io.Reader = ss.rsp.Body
	if debug {
		fmt.Fprintln(os.Stderr, ss.rsp)
		strm = io.TeeReader(strm, os.Stderr)
	}

	dec := NewDecoder(strm)
	for !ss.isStopped {
		rawMessage := make(map[string]json.RawMessage)
		err := dec.Decode(&rawMessage)
		if err != nil {
			// Server returned an ApiError instead of one of the documented Streaming JSON objects.
			// The likely reason is that the GET arguments were invalid so reconnecting won't
			// resolve the issue.  Therefore, stop the server before returning the error.
			if _, ok := err.(*ApiError); ok {
				ss.Stop()
				return err
			}

			// The likely failure is that the http.Response.Body was closed.  Ignore the error
			// and return from the method to reconnect if required.
			return nil
		}

		ss.stallTimer.Reset(ss.StallTimeout)

		err = ss.dispatchMessage(rawMessage, handleFn)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ss *streamServer) dispatchMessage(
	rawMessage map[string]json.RawMessage, handleFn messageHandleFunc) error {

	for msgType, msgData := range rawMessage {
		switch msgType {
		default:
			handleFn(msgType, msgData)

		case "heartbeat":
			hb := struct {
				Time time.Time `json:"time"`
			}{}
			if err := json.Unmarshal(msgData, &hb); err == nil {
				select {
				case ss.hbCh <- hb.Time:
				default:
				}
			}

		case "disconnect":
			// Notification that the server is about to disconnect.
			apiErr := ApiError{}
			if err := json.Unmarshal(msgData, &apiErr); err != nil {
				return err
			}
			return apiErr
		}
	}

	return nil
}

func (ss *streamServer) Stop() {
	ss.isStopped = true
	ss.disconnect()
}

func (ss *streamServer) init() error {
	if !ss.isStopped {
		return errors.New("Server is already running!")
	}
	ss.isStopped = false
	ss.stallTimer = time.AfterFunc(ss.StallTimeout, ss.connect)
	ss.startHeartbeatHandler()
	return nil
}

func (ss *streamServer) cleanup() {
	ss.stallTimer.Stop()
	ss.disconnect()
	close(ss.hbCh)
}

// connect issues a GET request and receives the http.Response object which is stores on the
// pricesServer instance.
func (ss *streamServer) connect() {
	ss.disconnect()

	var err error
	backoff := time.Second
	for !ss.isStopped {
		func() {
			ss.rspMtx.Lock()
			defer ss.rspMtx.Unlock()

			ss.rsp, err = ss.ctx.Connect()
			if err == nil {
				ss.stallTimer.Reset(ss.StallTimeout)
			}
		}()

		if err == nil {
			return
		}
		time.Sleep(backoff)
		backoff *= 2
	}
}

// disconnect closes the connection to the Oanda server.
func (ss *streamServer) disconnect() {
	ss.rspMtx.Lock()
	defer ss.rspMtx.Unlock()
	if ss.rsp != nil {
		ss.rsp.Body.Close()
	}
}

func (ss *streamServer) startHeartbeatHandler() {
	ss.wg.Add(1)
	go func() {
		var handleFn HeartbeatHandleFunc
		for hb := range ss.hbCh {
			// It's possible that the handle func on the pricesServer instance is changed or set
			// to nil, so take a copy.
			handleFn = ss.HeartbeatFunc
			if handleFn != nil {
				handleFn(hb)
			}
		}
		ss.wg.Done()
	}()
}
