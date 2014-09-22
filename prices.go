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

type Prices map[string]PriceTick

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
func (c *Client) PollPrices(instrument string, instruments ...string) (Prices, error) {
	return c.PollPricesSince(time.Time{}, instrument, instruments...)
}

// PollPricesSince returns the PriceTicks for instruments.  If since is not the zero time
// instruments whose prices were not updated since the requested time.Time are excluded from the
// result.
func (c *Client) PollPricesSince(since time.Time, instr string, instrs ...string) (Prices, error) {
	pp, err := c.NewPricePoller(since, instr, instrs...)
	if err != nil {
		return nil, err
	}
	return pp.Poll()
}

type PricePoller struct {
	pr         *PollRequest
	lastPrices Prices
}

// NewPricePoller returns a poller to effeciently poll Oanda for updates of the same set of
// instruments.
func (c *Client) NewPricePoller(since time.Time, instr string, instrs ...string) (*PricePoller, error) {
	instrs = append(instrs, instr)
	req, err := c.NewRequest("GET", "/v1/prices", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("instruments", strings.ToUpper(strings.Join(instrs, ",")))
	if !since.IsZero() {
		q.Set("since", since.UTC().Format(time.RFC3339))
	}
	req.URL.RawQuery = q.Encode()
	pp := PricePoller{
		pr:         &PollRequest{c, req},
		lastPrices: make(Prices),
	}
	return &pp, err
}

// Poll returns the most recent set of prices for the instruments with which the PricePoller
// was configured.
func (pp *PricePoller) Poll() (Prices, error) {
	rsp, err := pp.pr.Poll()
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.ContentLength == 0 {
		return pp.lastPrices, nil
	}

	v := struct {
		ApiError
		Prices []struct {
			Instrument string `json:"instrument"`
			PriceTick
		} `json:"prices"`
	}{}
	dec := NewDecoder(rsp.Body)
	if err = dec.Decode(&v); err != nil {
		return nil, err
	}
	prices := make(Prices)
	for _, p := range v.Prices {
		prices[p.Instrument] = p.PriceTick
	}
	pp.lastPrices = prices
	return prices, nil
}

type instrumentTick struct {
	Instrument string `json:"instrument"`
	PriceTick
}

var tickPool = sync.Pool{
	New: func() interface{} { return &instrumentTick{} },
}

type TickHandlerFunc func(instr string, pp PriceTick)

///////////////////////////////////////////////////////////////////////////////////////////////////
// priceServer

type priceServer struct {
	HeartbeatFunc HeartbeatHandlerFunc
	srv           *MessageServer
	chanMap       *tickChans
}

// NewPriceServer creates a Price Server for receiving and handling Ticks.
func (c *Client) NewPriceServer(instr string, instrs ...string) (*priceServer, error) {
	instrs = append(instrs, instr)
	for i, instr := range instrs {
		instrs[i] = strings.ToUpper(instr)
	}

	req, err := c.NewRequest("GET", "/v1/prices", nil)
	if err != nil {
		return nil, err
	}
	useStreamHost(req)

	u := req.URL
	q := u.Query()
	q.Set("instruments", strings.Join(instrs, ","))
	u.RawQuery = q.Encode()

	ps := priceServer{
		chanMap: newTickChans(instrs),
	}

	streamSrv := StreamServer{
		HandleMessageFn:   ps.handleMessage,
		HandleHeartbeatFn: ps.handleHeartbeat,
	}

	if srv, err := c.NewMessageServer(req, &streamSrv); err != nil {
		return nil, err
	} else {
		ps.srv = srv
	}

	return &ps, nil
}

// ConnectAndHandle connects to the Oanda server and invokes handleFn for every Tick received.
func (ps *priceServer) ConnectAndHandle(handleFn TickHandlerFunc) error {
	ps.initServer(handleFn)
	err := ps.srv.ConnectAndDispatch()
	ps.finish()
	return err
}

// Stop terminates the Price server.
func (ps *priceServer) Stop() {
	ps.srv.Stop()
}

func (ps *priceServer) initServer(handleFn TickHandlerFunc) {
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

func (ps *priceServer) finish() {
	for _, instr := range ps.chanMap.Instruments() {
		tickC, ok := ps.chanMap.Get(instr)
		if ok && tickC != nil {
			ps.chanMap.Set(instr, nil)
			close(tickC)
		}
	}
}

func (ps *priceServer) handleHeartbeat(hb time.Time) {
	if ps.HeartbeatFunc != nil {
		ps.HeartbeatFunc(hb)
	}
}

func (ps *priceServer) handleMessage(msgType string, rawMessage json.RawMessage) {
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
