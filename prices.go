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

type tickChans struct {
	mtx sync.RWMutex
	m   map[string]chan *instrumentTick
}

func (tc *tickChans) Instruments() []string {
	tc.mtx.RLock()
	defer tc.mtx.RUnlock()
	instruments := make([]string, len(tc.m))
	for instr := range tc.m {
		instruments = append(instruments, instr)
	}
	return instruments
}

func (tc *tickChans) Set(instr string, ch chan *instrumentTick) {
	tc.mtx.Lock()
	defer tc.mtx.Unlock()
	tc.m[instr] = ch
}

func (tc *tickChans) Get(instr string) (chan *instrumentTick, bool) {
	tc.mtx.RLock()
	defer tc.mtx.RUnlock()
	ch, ok := tc.m[instr]
	return ch, ok
}

func newTickChans(instruments []string) *tickChans {
	m := make(map[string]chan *instrumentTick)
	for _, instr := range instruments {
		m[instr] = nil
	}
	return &tickChans{
		m: m,
	}
}

var tickPool = sync.Pool{
	New: func() interface{} { return &instrumentTick{} },
}

type TickHandlerFunc func(instr string, pp PriceTick)

type StreamServer struct {
	HandleMessageFn   MessageHandlerFunc
	HandleHeartbeatFn HeartbeatHandlerFunc
}

func (pss StreamServer) HandleMessage(msgType string, msgData json.RawMessage) {
	if pss.HandleMessageFn != nil {
		pss.HandleMessageFn(msgType, msgData)
	}
}

func (pss StreamServer) HandleHeartbeat(hb time.Time) {
	if pss.HandleHeartbeatFn != nil {
		pss.HandleHeartbeatFn(hb)
	}
}

type pricesServer struct {
	HeartbeatFunc HeartbeatHandlerFunc
	srv           *server
	chanMap       *tickChans
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

	ps := pricesServer{
		chanMap: newTickChans(instruments),
	}

	streamSrv := StreamServer{
		HandleMessageFn:   ps.handleMessage,
		HandleHeartbeatFn: ps.handleHeartbeat,
	}

	if srv, err := c.NewServer(u, &streamSrv); err != nil {
		return nil, err
	} else {
		ps.srv = srv
	}

	return &ps, nil
}

func (ps *pricesServer) Run(handleFn TickHandlerFunc) error {
	ps.initServer(handleFn)
	err := ps.srv.Run()
	ps.finish()
	return err
}

func (ps *pricesServer) Stop() {
	ps.srv.Stop()
}

func (ps *pricesServer) initServer(handleFn TickHandlerFunc) {
	for _, instr := range ps.chanMap.Instruments() {
		tickC := make(chan *instrumentTick, defaultBufferSize)
		ps.chanMap.Set(instr, tickC)

		go func(lclC <-chan *instrumentTick) {
			for tick := range lclC {
				handleFn(tick.Instrument, tick.PriceTick)
				tickPool.Put(tick)
			}
		}(tickC)
	}
}

func (ps *pricesServer) finish() {
	for _, instr := range ps.chanMap.Instruments() {
		tickC, ok := ps.chanMap.Get(instr)
		if ok && tickC != nil {
			ps.chanMap.Set(instr, nil)
			close(tickC)
		}
	}
}

func (ps *pricesServer) handleHeartbeat(hb time.Time) {
	if ps.HeartbeatFunc != nil {
		ps.HeartbeatFunc(hb)
	}
}

func (ps *pricesServer) handleMessage(msgType string, rawMessage json.RawMessage) {
	tick := tickPool.Get().(*instrumentTick)
	if err := json.Unmarshal(rawMessage, tick); err != nil {
		ps.Stop()
		return
	}

	tickC, ok := ps.chanMap.Get(tick.Instrument)
	if !ok {
		// FIXME: Log error "unexpected instrument"
	} else if tickC != nil {
		tickC <- tick
	} else {
		// FIXME: Log "tick after server closed"
	}
}
