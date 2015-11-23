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
	"github.com/santegoeds/oanda"

	check "gopkg.in/check.v1"
)

type TestPositionSuite struct {
	OandaSuite
}

func (ts *TestPositionSuite) SetUpSuite(c *check.C) {
	ts.OandaSuite.SetUpSuite(c)
	ts.OandaSuite.SetUpAccount(c)
}

var _ = check.Suite(&TestPositionSuite{})

func (ts *TestPositionSuite) TestPositionsApi(c *check.C) {
	t, err := ts.Client.NewTrade(oanda.Buy, 1, "eur_usd")
	c.Assert(err, check.IsNil)
	c.Log(t)

	positions, err := ts.Client.Positions()
	c.Assert(err, check.IsNil)
	c.Log(positions)
	c.Assert(positions, check.HasLen, 1)
	c.Assert(positions[0].Side, check.Equals, t.Side)
	c.Assert(positions[0].Units, check.Equals, t.Units)
	c.Assert(positions[0].AvgPrice, check.Equals, t.Price)

	p, err := ts.Client.Position("eur_usd")
	c.Assert(err, check.IsNil)
	c.Log(p)
	c.Assert(p.Side, check.Equals, t.Side)
	c.Assert(p.Units, check.Equals, t.Units)
	c.Assert(p.AvgPrice, check.Equals, t.Price)

	cpr, err := ts.Client.ClosePosition("eur_usd")
	c.Assert(err, check.IsNil)
	c.Log(cpr)
	c.Assert(cpr.TranIds, check.HasLen, 2)
	c.Assert(cpr.TotalUnits, check.Equals, 1)
	c.Assert(cpr.Instrument, check.Equals, "EUR_USD")

	positions, err = ts.Client.Positions()
	c.Assert(err, check.IsNil)
	c.Assert(positions, check.HasLen, 0)
}

func (ts *TestPositionSuite) TestNonexistingPosition(c *check.C) {
	_, err := ts.Client.Position("eur_gbp")
	c.Assert(err, check.NotNil)

	apiErr, ok := err.(*oanda.ApiError)
	c.Assert(ok, check.Equals, true)
	c.Assert(apiErr.Code, check.Equals, 14)
}
