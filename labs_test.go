package oanda_test

import (
	"strings"

	"github.com/santegoeds/oanda"

	"gopkg.in/check.v1"
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
