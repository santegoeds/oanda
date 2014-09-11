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

	"github.com/santegoeds/oanda"

	"gopkg.in/check.v1"
)

func (ts *TestSuite) TestTradeApi(c *check.C) {
	t, err := ts.c.NewTrade(oanda.Ts_Buy, 2, "eur_usd", oanda.StopLoss(0.5), oanda.TakeProfit(3.0))
	if err != nil {
		c.Error(err)
		return
	}

	c.Log(t)

	c.Assert(t.TradeId, check.Not(check.Equals), 0)
	c.Assert(t.Price, check.Not(check.Equals), 0.0)
	c.Assert(t.Instrument, check.Equals, "EUR_USD")
	c.Assert(t.Side, check.Equals, string(oanda.Ts_Buy))
	c.Assert(t.Units, check.Equals, 2)
	c.Assert(t.StopLoss, check.Equals, 0.5)
	c.Assert(t.TakeProfit, check.Equals, 3.0)
	c.Assert(t.TrailingStop, check.Equals, 0.0)
	c.Assert(t.Time.Before(time.Now()), check.Equals, true)

	dup, err := ts.c.Trade(t.TradeId)
	if err != nil {
		c.Error(err)
		return
	}

	c.Assert(dup.TradeId, check.Equals, t.TradeId)
	c.Assert(dup.Price, check.Equals, t.Price)
	c.Assert(dup.Instrument, check.Equals, t.Instrument)
	c.Assert(dup.Side, check.Equals, t.Side)
	c.Assert(dup.Units, check.Equals, t.Units)
	c.Assert(dup.StopLoss, check.Equals, t.StopLoss)
	c.Assert(dup.TakeProfit, check.Equals, t.TakeProfit)
	c.Assert(dup.TrailingStop, check.Equals, t.TrailingStop)
	c.Assert(dup.Time.Equal(t.Time), check.Equals, true)

	t, err = ts.c.ModifyTrade(t.TradeId, oanda.StopLoss(0.75))
	if err != nil {
		c.Error(err)
		return
	}

	c.Assert(t.StopLoss, check.Equals, 0.75)

	trades, err := ts.c.Trades()
	if err != nil {
		c.Error(err)
		return
	}

	c.Assert(trades, check.HasLen, 1)
	c.Assert(trades[0].TradeId, check.Equals, t.TradeId)
	c.Assert(trades[0].Price, check.Equals, t.Price)
	c.Assert(trades[0].Instrument, check.Equals, t.Instrument)
	c.Assert(trades[0].Side, check.Equals, t.Side)
	c.Assert(trades[0].Units, check.Equals, t.Units)
	c.Assert(trades[0].StopLoss, check.Equals, t.StopLoss)
	c.Assert(trades[0].TakeProfit, check.Equals, t.TakeProfit)
	c.Assert(trades[0].TrailingStop, check.Equals, t.TrailingStop)
	c.Assert(trades[0].Time.Equal(t.Time), check.Equals, true)

	rsp, err := ts.c.CloseTrade(t.TradeId)
	if err != nil {
		c.Error(err)
		return
	}

	c.Log(rsp)

	trades, err = ts.c.Trades()
	if err != nil {
		c.Error(err)
		return
	}
	c.Assert(trades, check.HasLen, 0)
}
