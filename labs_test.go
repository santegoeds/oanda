package oanda_test

import (
	"os"

	"github.com/santegoeds/oanda"

	"gopkg.in/check.v1"
)

type TestLabsSuite struct {
	c *oanda.Client
}

var _ = check.Suite(&TestLabsSuite{})

func (ts *TestLabsSuite) SetUpSuite(c *check.C) {
	token := os.Getenv("FXPRACTICE_TOKEN")
	if token == "" {
		c.Skip("FXPRACTICE_TOKEN is not set")
	} else {
		client, err := oanda.NewFxPracticeClient(token)
		c.Assert(err, check.IsNil)
		ts.c = client
	}
}

func (ts *TestLabsSuite) TestCalendar(c *check.C) {
	events, err := ts.c.Calendar("eur_usd", oanda.Year)
	c.Assert(err, check.IsNil)
	c.Log(events)
}
