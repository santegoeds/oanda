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
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultBufferSize = 5
	maxDelay          = 5 * time.Minute
)

type (
	HeartbeatHandlerFunc  func(Time)
	messagesHandlerFunc   func(<-chan StreamMessage)
	heartbeatsHandlerFunc func(<-chan Time)
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

func (msg StreamMessage) String() string {
	return fmt.Sprintf("StreamMessage{%s, %s}", msg.Type, string(msg.RawMessage))
}

func (msg *StreamMessage) UnmarshalJSON(data []byte) error {
	msgMap := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &msgMap); err != nil {
		return err
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
	HandleHeartbeats(<-chan Time)
	HandleMessages(<-chan StreamMessage)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// StreamReader

type StreamServer struct {
	handleMessagesFn   messagesHandlerFunc
	handleHeartbeatsFn heartbeatsHandlerFunc
}

func (ss StreamServer) HandleMessages(msgC <-chan StreamMessage) {
	if ss.handleMessagesFn != nil {
		ss.handleMessagesFn(msgC)
	}
}

func (ss StreamServer) HandleHeartbeats(hbC <-chan Time) {
	if ss.handleHeartbeatsFn != nil {
		ss.handleHeartbeatsFn(hbC)
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// messageServer

type messageServer struct {
	sh           StreamHandler
	c            *Client
	mtx          sync.Mutex
	req          *http.Request
	runFlg       bool
	stallTimeout time.Duration
}

// newMessageServer returns a new instance of messageServer that forwards each message and
// heartbeat to the specified StreamHandler.
func (c *Client) newMessageServer(req *http.Request, sh StreamHandler, stallTimeout time.Duration) (*messageServer, error) {
	s := messageServer{
		sh:           sh,
		c:            c,
		req:          req,
		stallTimeout: stallTimeout,
	}
	return &s, nil
}

// ConnectAndDispatch
func (s *messageServer) ConnectAndDispatch() (err error) {
	if err = s.initServer(); err != nil {
		return
	}
	err = s.readMessages()

	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.runFlg = false
	return
}

// Stop stops the messageServer.
func (s *messageServer) Stop() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.runFlg = false
	cancelRequest(s)
}

func (s *messageServer) initServer() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if s.runFlg {
		return errors.New("server is already running")
	}
	s.runFlg = true
	return nil
}

func (s *messageServer) readMessages() error {
	hbC := make(chan Time)
	defer close(hbC)
	go s.sh.HandleHeartbeats(hbC)

	msgC := make(chan StreamMessage)
	defer close(msgC)
	go s.sh.HandleMessages(msgC)

	newResponse := func() (*http.Response, error) {
		rsp, err := s.c.Do(s.req)
		if err != nil {
			return nil, err
		}
		if rsp.StatusCode < 400 {
			return rsp, nil
		}
		apiErr := ApiError{}
		if err = json.NewDecoder(rsp.Body).Decode(&apiErr); err != nil {
			return nil, err
		}
		return nil, &apiErr
	}

	newReader := func() (rdr io.ReadCloser, err error) {
		delay := time.Second
		for {
			s.mtx.Lock()
			runFlg := s.runFlg
			if runFlg {
				var rsp *http.Response
				rsp, err = newResponse()
				if err != nil {
					_, ok := err.(*ApiError)
					runFlg = !ok
				} else {
					rdr = NewTimedReader(rsp.Body, s.stallTimeout)
				}
			}
			s.mtx.Unlock()
			if !runFlg || rdr != nil || delay >= maxDelay {
				break
			}
			time.Sleep(delay)
			delay *= 2
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
					Time Time `json:"time"`
				}{}
				if err := json.Unmarshal(msg.RawMessage, &v); err != nil {
					// FIXME: log error
				} else {
					hbC <- v.Time
				}
			case "disconnect":
				apiErr := ApiError{}
				if err = json.Unmarshal(msg.RawMessage, &apiErr); err == nil {
					err = &apiErr
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

func cancelRequest(s *messageServer) {
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
