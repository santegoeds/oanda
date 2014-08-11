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
	"time"

	"oanda"

	"gopkg.in/check.v1"
)

func (ts *TestSuite) TestPollPrices(c *check.C) {
	prices, err := ts.c.PollPrices("EUR_USD", "EUR_GBP")
	if err != nil {
		c.Error(err)
		return
	}

	c.Log(prices)

	c.Assert(prices, check.HasLen, 2)

	for _, pi := range prices {
		c.Assert(pi.Spread() > 0.0, check.Equals, true)
	}
}

func (ts *TestSuite) TestPollPricesSince(c *check.C) {
	prices, err := ts.c.PollPricesSince(time.Now().Add(-time.Hour), "eur_usd")
	if err != nil {
		c.Error(err)
		return
	}

	c.Assert(prices, check.HasLen, 1)
}

func (ts *TestSuite) TestPollPricesContext(c *check.C) {
	ctx, err := ts.c.NewPollPricesContext(time.Time{}, "eur_usd", "eur_gbp")
	if err != nil {
		c.Error(err)
		return
	}

	prices, err := ctx.Poll()
	if err != nil {
		c.Error(err)
		return
	}
	c.Log(prices)
	c.Assert(prices, check.HasLen, 2)

	prices, err = ctx.Poll()
	if err != nil {
		c.Error(err)
		return
	}
	c.Log(prices)
	c.Assert(prices, check.HasLen, 2)
}

func (ts *TestSuite) TestPricesServer(c *check.C) {
	ps, err := ts.c.NewPricesServer("eur_usd", "eur_gbp")
	if err != nil {
		c.Error(err)
		return
	}

	count := 0
	err = ps.Run(func(instrument string, tick oanda.PriceTick) {
		c.Log(instrument, tick)

		count += 1
		if count > 3 {
			ps.Stop()
		}
	})
	if err != nil {
		c.Error(err)
	}
}

func (ts *TestSuite) TestPricesServerInvalidInstrument(c *check.C) {
	ps, err := ts.c.NewPricesServer("gbp_eur")
	if err != nil {
		c.Error(err)
		return
	}

	err = ps.Run(func(in string, tick oanda.PriceTick) {
		c.Fail()
	})
	c.Assert(err, check.Not(check.IsNil))

	c.Log(err)
}
