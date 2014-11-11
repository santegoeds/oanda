package analytics_test

import (
	"testing"

	"gopkg.in/check.v1"

	"github.com/santegoeds/oanda/analytics"
)

type TestSuite struct{}

var _ = check.Suite(&TestSuite{})

func Test(t *testing.T) { check.TestingT(t) }

func (ts *TestSuite) TestWindow(c *check.C) {
	w := analytics.NewWindow(3)
	c.Assert(w.Cap(), check.Equals, 3)
	c.Assert(w.Len(), check.Equals, 0)

	w.Push(1.0)
	c.Assert(w.Cap(), check.Equals, 3)
	c.Assert(w.Len(), check.Equals, 1)

	w.Push(2.0, 3.0)
	c.Assert(w.Cap(), check.Equals, 3)
	c.Assert(w.Len(), check.Equals, 3)

	for i, v := range []float64{3, 2, 1} {
		c.Assert(v, check.Equals, w.Values()[i])
	}

	w.Push(4.0)
	for i, v := range []float64{4, 3, 2} {
		c.Assert(v, check.Equals, w.Values()[i])
	}

	w.Push(5, 6, 7, 8)
	for i, v := range []float64{8, 7, 6} {
		c.Assert(v, check.Equals, w.Values()[i])
	}
}
