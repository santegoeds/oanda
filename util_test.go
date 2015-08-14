package oanda_test

import (
	"time"

	"github.com/santegoeds/oanda"

	"gopkg.in/check.v1"
)

type UtilSuite struct {
	Time oanda.Time
}

var _ = check.Suite(&UtilSuite{})

func (s *UtilSuite) SetUpSuite(c *check.C) {
	s.Time = oanda.Time("1439662384000000")
}

func (s *UtilSuite) TestTimeUnixMicro(c *check.C) {
	c.Assert(s.Time.UnixMicro(), check.Equals, int64(1439662384000000))
}

func (s *UtilSuite) TestTimeUnixNano(c *check.C) {
	c.Assert(s.Time.UnixNano(), check.Equals, int64(1439662384000000000))
}

func (s *UtilSuite) TestTimeIsZero(c *check.C) {
	c.Assert(s.Time.IsZero(), check.Equals, false)

	zeroTime := oanda.Time("")
	c.Assert(zeroTime.IsZero(), check.Equals, true)
}

func (s *UtilSuite) TestTimeTime(c *check.C) {
	expected := time.Unix(0, 1439662384000000000)

	c.Assert(s.Time.Time(), check.Equals, expected)
}
