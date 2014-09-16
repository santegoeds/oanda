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
	"fmt"
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
		ApiError
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

var tickPool = sync.Pool{
	New: func() interface{} { return &instrumentTick{} },
}

type TickHandleFunc func(instrument string, pp PriceTick)

type pricesServer struct {
	BufferSize int

	*streamServer
	instruments []string
	tickChs     map[string]chan *instrumentTick
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

	ss, err := c.newStreamServer(u)
	if err != nil {
		return nil, err
	}

	ps := pricesServer{
		BufferSize:   DefaultBufferSize,
		streamServer: ss,
		instruments:  instruments,
		tickChs:      make(map[string]chan *instrumentTick, len(instruments)),
	}

	return &ps, nil
}

// Run connects to the oanda server and dispatches PriceTicks to handleFn. A separate handleFun
// go-routine is started for each of the instruments.
func (ps *pricesServer) Run(handleFn TickHandleFunc) error {
	err := ps.init(handleFn)
	if err != nil {
		return err
	}
	defer ps.cleanup()

	err = ps.streamServer.Run(func(msgType string, msgData json.RawMessage) error {
		if msgType != "tick" {
			return fmt.Errorf("%s is an unexpected message type", msgType)
		}
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

		return nil
	})

	return err
}

func (ps *pricesServer) init(handleFn TickHandleFunc) error {
	for _, in := range ps.instruments {
		ps.tickChs[in] = make(chan *instrumentTick)
	}
	for _, ch := range ps.tickChs {
		ps.wg.Add(1)
		go func(tickCh <-chan *instrumentTick) {
			for tick := range tickCh {
				handleFn(tick.Instrument, tick.PriceTick)
				tickPool.Put(tick)
			}
			ps.wg.Done()
		}(ch)
	}

	return nil
}

func (ps *pricesServer) cleanup() {
	for _, tickCh := range ps.tickChs {
		close(tickCh)
	}
	ps.wg.Wait()
}
