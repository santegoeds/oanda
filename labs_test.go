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

func (ts *TestLabsSuite) TestLabsHistoricPositionRatios(c *check.C) {
	instrument := "eur_usd"
	ratios, err := ts.c.HistoricPositionRatios("eur_usd", oanda.Year)
	c.Assert(err, check.IsNil)
	c.Log(ratios)
	instrument = strings.ToUpper(instrument)
	c.Assert(ratios.Instrument, check.Equals, instrument)
	c.Assert(ratios.DisplayName, check.Equals, strings.Replace(instrument, "_", "/", -1))
	c.Assert(len(ratios.Ratios) > 0, check.Equals, true)
}
