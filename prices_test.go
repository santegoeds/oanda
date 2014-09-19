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
package oanda_test

import (
	"sync"
	"time"

	"github.com/santegoeds/oanda"

	"gopkg.in/check.v1"
)

type Counter struct {
	m sync.RWMutex
	n int
}

func (c *Counter) Inc() int {
	c.m.Lock()
	defer c.m.Unlock()
	c.n++
	return c.n
}

func (c *Counter) Val() int {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.n
}

func (ts *TestSuite) TestPollPrices(c *check.C) {
	prices, err := ts.c.PollPrices("EUR_USD", "EUR_GBP")
	c.Assert(err, check.IsNil)
	c.Log(prices)
	c.Assert(prices, check.HasLen, 2)

	for _, pi := range prices {
		c.Assert(pi.Spread() > 0.0, check.Equals, true)
	}
}

func (ts *TestSuite) TestPollPricesSince(c *check.C) {
	prices, err := ts.c.PollPricesSince(time.Now().Add(-time.Hour), "eur_usd")
	c.Assert(err, check.IsNil)
	c.Assert(prices, check.HasLen, 1)
}

func (ts *TestSuite) TestPollPricesContext(c *check.C) {
	ctx, err := ts.c.NewPollPricesContext(time.Time{}, "eur_usd", "eur_gbp")
	c.Assert(err, check.IsNil)

	prices, err := ctx.Poll()
	c.Assert(err, check.IsNil)
	c.Log(prices)
	c.Assert(prices, check.HasLen, 2)

	prices, err = ctx.Poll()
	c.Assert(err, check.IsNil)
	c.Log(prices)
	c.Assert(prices, check.HasLen, 2)
}

func (ts *TestSuite) TestPricesServer(c *check.C) {
	ps, err := ts.c.NewPricesServer("eur_usd", "eur_gbp")
	c.Assert(err, check.IsNil)

	timeout := 2 * time.Minute
	t := time.AfterFunc(timeout, func() {
		c.Errorf("Failed to receive 3 ticks in %d minutes.", timeout/time.Minute)
		ps.Stop()
	})

	count := Counter{}
	err = ps.Run(func(instrument string, tick oanda.PriceTick) {
		c.Log(instrument, tick)
		if count.Inc() > 3 {
			ps.Stop()
		}
	})

	t.Stop()
	c.Assert(err, check.IsNil)
	c.Assert(count.Val() > 3, check.Equals, true)
}

func (ts *TestSuite) TestPricesServerInvalidInstrument(c *check.C) {
	ps, err := ts.c.NewPricesServer("gbp_eur")
	c.Assert(err, check.IsNil)
	err = ps.Run(func(in string, tick oanda.PriceTick) {
		c.Fail()
	})
	c.Assert(err, check.NotNil)
	c.Log(err)
}

func (ts *TestSuite) TestPricesServerHeartbeat(c *check.C) {
	ps, err := ts.c.NewPricesServer("gbp_aud")
	c.Assert(err, check.IsNil)

	ps.HeartbeatFunc = func(hb time.Time) {
		c.Logf("Heartbeat: %s", hb)
		ps.Stop()
	}

	err = ps.Run(func(in string, tick oanda.PriceTick) {
		c.Log(in, tick)
	})
	c.Assert(err, check.IsNil)
}
