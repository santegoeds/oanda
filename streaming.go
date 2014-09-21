// Copyright 2014 Tjerk Santegoeds
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package oanda

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultBufferSize   = 5
	defaultStallTimeout = 10 * time.Second
	maxDelay            = 5 * time.Minute
)

type (
	MessageHandlerFunc   func(string, json.RawMessage)
	HeartbeatHandlerFunc func(time.Time)
)

///////////////////////////////////////////////////////////////////////////////////////////////////
// TimedReader

type TimedReader struct {
	Timeout time.Duration
	io.ReadCloser
	timer *time.Timer
}

// NewTimedReader returns an instance of TimedReader where Read operations time out.
func NewTimedReader(r io.ReadCloser, timeout time.Duration) *TimedReader {
	return &TimedReader{
		Timeout:    timeout,
		ReadCloser: r,
	}
}

func (r *TimedReader) Read(p []byte) (int, error) {
	if r.timer == nil {
		r.timer = time.AfterFunc(r.Timeout, func() { r.Close() })
	} else {
		r.timer.Reset(r.Timeout)
	}
	n, err := r.ReadCloser.Read(p)
	r.timer.Stop()
	return n, err
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// StreamMessage

type StreamMessage struct {
	Type       string
	RawMessage json.RawMessage
}

func (msg *StreamMessage) UnmarshalJSON(data []byte) error {
	msgMap := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &msgMap); err != nil {
		return err
	}
	if code, ok := msgMap["code"]; ok {
		apiError := ApiError{}
		if err := json.Unmarshal(code, &apiError.Code); err != nil {
			return err
		}
		if apiError.Code != 0 {
			json.Unmarshal(msgMap["message"], &apiError.Message)
			json.Unmarshal(msgMap["moreInfo"], &apiError.MoreInfo)
			return &apiError
		}
	}

	for msgType, rawMessage := range msgMap {
		msg.Type = msgType
		msg.RawMessage = rawMessage
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// StreamHandler

type StreamHandler interface {
	HandleHeartbeat(time.Time)
	HandleMessage(msgType string, rawMessage json.RawMessage)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// StreamReader

type StreamServer struct {
	HandleMessageFn   MessageHandlerFunc
	HandleHeartbeatFn HeartbeatHandlerFunc
}

func (ss StreamServer) HandleMessage(msgType string, msgData json.RawMessage) {
	if ss.HandleMessageFn != nil {
		ss.HandleMessageFn(msgType, msgData)
	}
}

func (ss StreamServer) HandleHeartbeat(hb time.Time) {
	if ss.HandleHeartbeatFn != nil {
		ss.HandleHeartbeatFn(hb)
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// MessageServer

type MessageServer struct {
	sh     StreamHandler
	c      *Client
	mtx    sync.Mutex
	req    *http.Request
	runFlg bool
}

// NewMessageServer returns a new instance of MessageServer that forwards each message and
// heartbeat to the specified StreamHandler.
func (c *Client) NewMessageServer(req *http.Request, sh StreamHandler) (*MessageServer, error) {
	s := MessageServer{
		sh:  sh,
		c:   c,
		req: req,
	}
	return &s, nil
}

// ConnectAndDispatch
func (s *MessageServer) ConnectAndDispatch() (err error) {
	if err = s.initServer(); err != nil {
		return
	}
	err = s.readMessages()

	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.runFlg = false
	return
}

// Stop stops the MessageServer.
func (s *MessageServer) Stop() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.runFlg = false
	cancelRequest(s)
}

func (s *MessageServer) initServer() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if s.runFlg {
		return errors.New("server is already running")
	}
	s.runFlg = true
	return nil
}

func (s *MessageServer) readMessages() error {
	hbC := make(chan time.Time)
	defer close(hbC)
	go func() {
		for hb := range hbC {
			s.sh.HandleHeartbeat(hb)
		}
	}()

	msgC := make(chan StreamMessage)
	defer close(msgC)
	go func() {
		for msg := range msgC {
			s.sh.HandleMessage(msg.Type, msg.RawMessage)
		}
	}()

	newReader := func() (rdr io.ReadCloser, err error) {
		d := time.Second
		for {
			s.mtx.Lock()
			runFlg := s.runFlg
			if runFlg {
				rsp, err := s.c.Do(s.req)
				if err == nil {
					rdr = NewTimedReader(rsp.Body, defaultStallTimeout)
				}
			}
			s.mtx.Unlock()
			if !runFlg || rdr != nil || d >= maxDelay {
				break
			}
			time.Sleep(d)
			d *= 2
		}
		return
	}

	for {
		rdr, err := newReader()
		if rdr == nil || err != nil {
			return err
		}
		dec := json.NewDecoder(rdr)

		msg := StreamMessage{}
		for {
			err = dec.Decode(&msg)
			if err != nil {
				if _, ok := err.(*ApiError); ok {
					rdr.Close()
					return err
				}
				break
			}

			switch msg.Type {
			default:
				msgC <- msg
			case "heartbeat":
				v := struct {
					Time time.Time `json:"time"`
				}{}
				if err := json.Unmarshal(msg.RawMessage, &v); err != nil {
					// FIXME: log error
				} else {
					hbC <- v.Time
				}
			case "disconnect":
				apiErr := ApiError{}
				if err = json.Unmarshal(msg.RawMessage, &apiErr); err == nil {
					err = apiErr
				}
				// FIXME: log msg.AsApiError()
				s.mtx.Lock()
				cancelRequest(s)
				s.mtx.Unlock()
				break
			}
		}
		rdr.Close()
	}
}

func cancelRequest(s *MessageServer) {
	if s.req != nil {
		s.c.CancelRequest(s.req)
	}
}

func useStreamHost(req *http.Request) {
	u := req.URL
	parts := strings.Split(u.Host, "-")
	parts[0] = "stream"
	u.Host = strings.Join(parts, "-")
}
