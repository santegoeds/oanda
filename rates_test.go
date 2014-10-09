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
	"strings"

	"gopkg.in/check.v1"

	"github.com/santegoeds/oanda"
)

type TestRatesSuite struct {
	c *oanda.Client
}

var _ = check.Suite(&TestRatesSuite{})

func (ts *TestRatesSuite) SetUpSuite(c *check.C) {
	ts.c = NewTestClient(c, true)
}

func (ts *TestRatesSuite) TestRatesInstruments(c *check.C) {
	instruments, err := ts.c.Instruments(nil, nil)
	c.Assert(err, check.IsNil)
	c.Log(instruments)
	c.Assert(instruments, check.Not(check.HasLen), 0)
}

func (ts *TestRatesSuite) TestRatesMidpointCandles(c *check.C) {
	instrument, granularity := "eur_usd", oanda.D
	candles, err := ts.c.PollMidpointCandles(instrument, granularity)
	c.Assert(err, check.IsNil)
	c.Log(candles)
	c.Assert(candles.Instrument, check.Equals, strings.ToUpper(instrument))
	c.Assert(candles.Granularity, check.Equals, granularity)
	c.Assert(len(candles.Candles) > 0, check.Equals, true)
}

func (ts *TestRatesSuite) TestRatesBidAskCandles(c *check.C) {
	instrument, granularity := "eur_usd", oanda.D
	candles, err := ts.c.PollBidAskCandles(instrument, granularity)
	c.Assert(err, check.IsNil)
	c.Log(candles)
	c.Assert(candles.Instrument, check.Equals, strings.ToUpper(instrument))
	c.Assert(candles.Granularity, check.Equals, granularity)
	c.Assert(len(candles.Candles) > 0, check.Equals, true)
}
