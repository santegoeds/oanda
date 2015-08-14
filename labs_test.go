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

	"github.com/santegoeds/oanda"
	check "gopkg.in/check.v1"
)

type TestLabsSuite struct {
	c *oanda.Client
}

var _ = check.Suite(&TestLabsSuite{})

func (ts *TestLabsSuite) SetUpSuite(c *check.C) {
	ts.c = NewTestClient(c, false)
}

func (ts *TestLabsSuite) TestLabsCalendar(c *check.C) {
	events, err := ts.c.Calendar("eur_usd", oanda.Year)
	c.Assert(err, check.IsNil)
	c.Log(events)
	c.Assert(len(events) > 0, check.Equals, true)
}

func (ts *TestLabsSuite) TestLabsPositionRatios(c *check.C) {
	instrument := "eur_usd"
	ratios, err := ts.c.PositionRatios(instrument, oanda.Year)
	c.Assert(err, check.IsNil)
	c.Log(ratios)
	instrument = strings.ToUpper(instrument)
	c.Assert(ratios.Instrument, check.Equals, instrument)
	c.Assert(ratios.DisplayName, check.Equals, strings.Replace(instrument, "_", "/", -1))
	c.Assert(len(ratios.Ratios) > 0, check.Equals, true)
}

func (ts *TestLabsSuite) TestLabsSpreads(c *check.C) {
	instrument := "eur_usd"
	spreads, err := ts.c.Spreads(instrument, oanda.Day, true)
	c.Assert(err, check.IsNil)
	c.Log(spreads)
	c.Assert(len(spreads.Max) > 0, check.Equals, true)
	c.Assert(len(spreads.Avg) > 0, check.Equals, true)
	c.Assert(len(spreads.Min) > 0, check.Equals, true)
}

func (ts *TestLabsSuite) TestLabsCommitmentsOfTraders(c *check.C) {
	instrument := "eur_usd"
	cot, err := ts.c.CommitmentsOfTraders(instrument)
	c.Assert(err, check.IsNil)
	c.Log(cot)
	c.Assert(len(cot) > 0, check.Equals, true)
}

func (ts *TestLabsSuite) TestLabsOrderBooks(c *check.C) {
	instrument, period := "eur_usd", 6*oanda.Hour
	obs, err := ts.c.OrderBooks(instrument, period)
	c.Assert(err, check.IsNil)
	c.Log(obs)
	c.Assert(len(obs) > 1, check.Equals, true)
	c.Assert(obs[0].Timestamp.UnixMicro() < obs[1].Timestamp.UnixMicro(), check.Equals, true)
}

func (ts *TestLabsSuite) TestLabsAutochartistPattern(c *check.C) {
	p, err := ts.c.AutochartistPattern()
	c.Assert(err, check.IsNil)
	c.Log(p)
	c.Assert(p.Provider, check.Equals, "autochartist")
	c.Assert(len(p.Signals) > 0, check.Equals, true)
}
