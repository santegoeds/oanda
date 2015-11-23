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
	OandaSuite
}

var _ = check.Suite(&TestRatesSuite{})

func (ts *TestRatesSuite) SetUpSuite(c *check.C) {
	ts.OandaSuite.SetUpSuite(c)
	ts.SetUpAccount(c)
}

func (ts *TestRatesSuite) TestRatesInstruments(c *check.C) {
	instruments, err := ts.Client.Instruments(nil, nil)
	c.Assert(err, check.IsNil)
	c.Log(instruments)
	c.Assert(instruments, check.Not(check.HasLen), 0)
}

func (ts *TestRatesSuite) TestRatesInstrumentsWithFields(c *check.C) {
	fields := []oanda.InstrumentField{
		oanda.DisplayNameField,
		oanda.PipField,
		oanda.MaxTradeUnitsField,
		oanda.PrecisionField,
		oanda.MaxTrailingStopField,
		oanda.MinTrailingStopField,
		oanda.MarginRateField,
		oanda.HaltedField,
		oanda.InterestRateField,
	}

	instruments, err := ts.Client.Instruments(nil, fields)
	c.Assert(err, check.IsNil)
	c.Log(instruments)
	c.Assert(instruments, check.Not(check.HasLen), 0)
	for _, info := range instruments {
		c.Assert(len(info.DisplayName) > 0, check.Equals, true)
		c.Assert(info.Pip > 0, check.Equals, true)
		c.Assert(info.MaxTradeUnits > 0, check.Equals, true)
		c.Assert(info.Precision > 0, check.Equals, true)
		c.Assert(info.MaxTrailingStop > 0, check.Equals, true)
		c.Assert(info.MinTrailingStop > 0, check.Equals, true)
		c.Assert(info.MarginRate > 0, check.Equals, true)
		// FIXME: Disable check as it regularly causes test failures when a market happens to
		// be closed
		//c.Assert(info.Halted, check.Equals, false)
		c.Assert(len(info.InterestRate) > 0, check.Equals, true)
	}
}

func (ts *TestRatesSuite) TestRatesMidpointCandles(c *check.C) {
	instrument, granularity := "eur_usd", oanda.D
	candles, err := ts.Client.PollMidpointCandles(instrument, granularity)
	c.Assert(err, check.IsNil)
	c.Log(candles)
	c.Assert(candles.Instrument, check.Equals, strings.ToUpper(instrument))
	c.Assert(candles.Granularity, check.Equals, granularity)
	c.Assert(len(candles.Candles) > 0, check.Equals, true)
}

func (ts *TestRatesSuite) TestRatesBidAskCandles(c *check.C) {
	instrument, granularity := "eur_usd", oanda.D
	candles, err := ts.Client.PollBidAskCandles(instrument, granularity)
	c.Assert(err, check.IsNil)
	c.Log(candles)
	c.Assert(candles.Instrument, check.Equals, strings.ToUpper(instrument))
	c.Assert(candles.Granularity, check.Equals, granularity)
	c.Assert(len(candles.Candles) > 0, check.Equals, true)
}
