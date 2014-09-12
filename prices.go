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
	"os"
	"strings"
	"sync"
	"time"
)

// PriceTick holds the Bid price, Ask price and status for an instrument at a given point
// in time
type PriceTick struct {
	Time   time.Time `json:"time"`
	Bid    float64   `json:"bid"`
	Ask    float64   `json:"ask"`
	Status string    `json:"status"`
}

// Spread returns the difference between Ask and Bid prices.
func (p *PriceTick) Spread() float64 {
	return p.Ask - p.Bid
}

// PollPrices returns the latest PriceTick for instruments.
func (c *Client) PollPrices(instrument string, instruments ...string) (
	map[string]PriceTick, error) {

	return c.PollPricesSince(time.Time{}, instrument, instruments...)
}

// PollPricesSince returns the PriceTicks for instruments.  If since is not the zero time
// instruments whose prices were not updated since the requested time.Time are excluded from the
// result.
func (c *Client) PollPricesSince(since time.Time, instrument string, instruments ...string) (
	map[string]PriceTick, error) {

	ctx, err := c.NewPollPricesContext(since, instrument, instruments...)
	if err != nil {
		return nil, err
	}
	return ctx.Poll()
}

type PollPricesContext struct {
	ctx *Context
}

func (ppc *PollPricesContext) Poll() (map[string]PriceTick, error) {
	v := struct {
		Prices []struct {
			Instrument string `json:"instrument"`
			PriceTick
		} `json:"prices"`
	}{}

	if _, err := ppc.ctx.Decode(&v); err != nil {
		return nil, err
	}

	prices := make(map[string]PriceTick)
	for _, p := range v.Prices {
		prices[p.Instrument] = p.PriceTick
	}
	return prices, nil
}

// NewPollPricesContext creates a context to repeatedly poll for PriceTicks using the same
// args.
func (c *Client) NewPollPricesContext(since time.Time, instrument string, instruments ...string) (
	*PollPricesContext, error) {

	instruments = append(instruments, instrument)

	u := c.getUrl("/v1/prices", "api")
	q := u.Query()
	q.Set("instruments", strings.ToUpper(strings.Join(instruments, ",")))
	if !since.IsZero() {
		q.Set("since", since.UTC().Format(time.RFC3339))
	}
	u.RawQuery = q.Encode()

	ctx, err := c.newContext("GET", u, nil)
	if err != nil {
		return nil, err
	}
	return &PollPricesContext{ctx}, nil
}

type instrumentTick struct {
	Instrument string `json:"instrument"`
	PriceTick
}

const (
	DefaultChannelSize  = 5
	DefaultStallTimeout = 10 * time.Second
)

var tickPool = sync.Pool{
	New: func() interface{} { return &instrumentTick{} },
}

type TickHandleFunc func(instrument string, pp PriceTick)

type pricesServer struct {
	ChannelSize  int
	StallTimeout time.Duration

	ctx         *Context
	instruments []string
	tickChs     map[string]chan *instrumentTick
	stallTimer  *time.Timer
	rsp         *http.Response
	rspMtx      sync.Mutex
	isStopped   bool
}

// NewPricesServer creates a pricesServer to receive and handle PriceTicks from the Oanda server.
func (c *Client) NewPricesServer(instrument string, instruments ...string) (*pricesServer, error) {
	instruments = append(instruments, instrument)
	for i := range instruments {
		instruments[i] = strings.ToUpper(instruments[i])
	}

	u := c.getUrl("/v1/prices", "stream")
	q := u.Query()
	q.Set("instruments", strings.Join(instruments, ","))
	u.RawQuery = q.Encode()

	ctx, err := c.newContext("GET", u, nil)
	if err != nil {
		return nil, err
	}

	ps := pricesServer{
		ChannelSize:  DefaultChannelSize,
		StallTimeout: DefaultStallTimeout,
		ctx:          ctx,
		instruments:  instruments,
		isStopped:    true,
	}

	return &ps, nil
}

// Run connects to the oanda server and dispatches PriceTicks to handleFn. A separate handleFun
// go-routine is started for each of the instruments.
func (ps *pricesServer) Run(handleFn TickHandleFunc) error {
	err := ps.init()
	if err != nil {
		return err
	}
	defer ps.cleanup()
	ps.startTickHandlers(handleFn)
	for !ps.isStopped {
		ps.connect()
		err = ps.dispatchTicks()
	}
	return err
}

// Stop terminates the connection to the oanda server.
func (ps *pricesServer) Stop() {
	ps.isStopped = true
	ps.disconnect()
}

func (ps *pricesServer) init() error {
	if !ps.isStopped {
		return errors.New("Server is already running!")
	}
	ps.isStopped = false
	ps.stallTimer = time.AfterFunc(ps.StallTimeout, ps.disconnect)
	ps.tickChs = make(map[string]chan *instrumentTick, ps.ChannelSize)
	for _, in := range ps.instruments {
		ps.tickChs[in] = make(chan *instrumentTick)
	}
	return nil
}

// cleanup makes sure that any timers are stopped and that the connection to the Oanda server
// is closed.
func (ps *pricesServer) cleanup() {
	ps.stallTimer.Stop()
	ps.disconnect()
	for _, tickCh := range ps.tickChs {
		close(tickCh)
	}
}

// startTickHanders starts one go-routine for each requested instrument.
func (ps *pricesServer) startTickHandlers(handleFn TickHandleFunc) {
	for _, ch := range ps.tickChs {
		go func(tickCh <-chan *instrumentTick) {
			for tick := range tickCh {
				handleFn(tick.Instrument, tick.PriceTick)
				tickPool.Put(tick)
			}
		}(ch)
	}
}

// connect issues a GET request and receives the http.Response object which is stores on the
// pricesServer instance.
func (ps *pricesServer) connect() {
	ps.disconnect()
	var err error
	backoff := time.Second
	for !ps.isStopped {
		func() {
			ps.rspMtx.Lock()
			defer ps.rspMtx.Unlock()

			ps.rsp, err = ps.ctx.Connect()
			if err == nil {
				ps.stallTimer.Reset(ps.StallTimeout)
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
func (ps *pricesServer) disconnect() {
	ps.rspMtx.Lock()
	defer ps.rspMtx.Unlock()
	if ps.rsp != nil {
		ps.rsp.Body.Close()
	}
}

// dispatchTicks reads the chunked response from the Oanda server, decodes ticks and dispatches
// them to the apropriate handler function.
func (ps *pricesServer) dispatchTicks() error {
	var strm io.Reader = ps.rsp.Body
	if debug {
		fmt.Fprintln(os.Stderr, ps.rsp)
		strm = io.TeeReader(strm, os.Stderr)
	}

	dec := json.NewDecoder(strm)
	for !ps.isStopped {
		rawMessage := make(map[string]json.RawMessage)
		err := dec.Decode(&rawMessage)
		if err != nil {
			// Likely failure is because the response stream is closed; an expected error if
			// the pricesServer has been stopped.
			return nil
		}

		ps.stallTimer.Reset(ps.StallTimeout)

		msgData, ok := rawMessage["tick"]
		if ok {

			// Dispatch Tick.
			tick := tickPool.Get().(*instrumentTick)
			if err = json.Unmarshal(msgData, tick); err != nil {
				return err
			}

			tickCh := ps.tickChs[tick.Instrument]
			select {
			case tickCh <- tick:
				// Nop
			default:
				// Channel is full. Remove a tick from the channel to make space.
				select {
				case <-tickCh:
				default:
				}
				tickCh <- tick
			}
		} else if msgData, ok = rawMessage["heartbeat"]; ok {

			// No need to process heartbeats.

		} else if msgData, ok = rawMessage["disconnect"]; ok {

			// Notification that the server is about to disconnect.
			apiError := ApiError{}
			if err = json.Unmarshal(msgData, &apiError); err != nil {
				return err
			}
			return apiError
		} else if msgData, ok = rawMessage["code"]; ok {

			// The Oanda server returned an error in a non-streaming format.  This is likely the
			// result of an invalid request parameter so reconnecting will not resolve the error.
			// The server is therefore stopped before the error is returned.
			apiError := ApiError{}
			if err = json.Unmarshal(msgData, &apiError.Code); err != nil {
				return err
			}
			if msgData, ok = rawMessage["message"]; ok {
				if err = json.Unmarshal(msgData, &apiError.Message); err != nil {
					return err
				}
			}
			if msgData, ok = rawMessage["moreInfo"]; ok {
				if err = json.Unmarshal(msgData, &apiError.MoreInfo); err != nil {
					return err
				}
			}
			ps.Stop()
			return apiError
		}
	}
	return nil
}
