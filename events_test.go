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
	"sync"
	"time"

	"github.com/santegoeds/oanda"

	"gopkg.in/check.v1"
)

type TestEventSuite struct {
	AccountId int
	c         *oanda.Client
}

var _ = check.Suite(&TestEventSuite{})

func (ts *TestEventSuite) SetUpSuite(c *check.C) {
	ts.c = NewTestClient(c, true)
}

func (ts *TestEventSuite) TearDownSuite(c *check.C) {
	CancelAllOrders(c, ts.c)
}

func (ts *TestEventSuite) TestEventApi(c *check.C) {
	expiry := time.Now().Add(24 * time.Hour)
	_, err := ts.c.NewOrder(oanda.Limit, oanda.Buy, 1, "eur_usd", 0.75, expiry)
	c.Assert(err, check.IsNil)

	events, err := ts.c.PollEvents(oanda.Count(1))
	c.Assert(err, check.IsNil)
	c.Log(events)
	c.Assert(events, check.HasLen, 1)

	c.Assert(events[0].AccountId(), check.Equals, ts.c.AccountId())
	c.Assert(events[0].Type(), check.Equals, "LIMIT_ORDER_CREATE")

	orderCreate1, ok := events[0].(*oanda.OrderCreateEvent)
	c.Assert(ok, check.Equals, true)
	c.Assert(orderCreate1.Instrument(), check.Equals, "EUR_USD")
	c.Assert(orderCreate1.Side(), check.Equals, "buy")
	c.Assert(orderCreate1.Units(), check.Equals, 1)
	c.Assert(orderCreate1.Price(), check.Equals, 0.75)

	orderExpiry := orderCreate1.Expiry().Time()
	c.Assert(orderExpiry.Equal(expiry.Truncate(time.Second)), check.Equals, true)
	c.Assert(orderCreate1.Reason(), check.Equals, "CLIENT_REQUEST")

	evt, err := ts.c.PollEvent(orderCreate1.TranId())
	c.Assert(err, check.IsNil)

	c.Log(evt)
	c.Check(evt.Type(), check.Equals, orderCreate1.Type())
	c.Check(evt.AccountId(), check.Equals, orderCreate1.AccountId())
	c.Check(evt.Time(), check.Equals, orderCreate1.Time())

	orderCreate2, ok := evt.(*oanda.OrderCreateEvent)
	c.Assert(ok, check.Equals, true)

	c.Assert(orderCreate2.Instrument(), check.Equals, orderCreate1.Instrument())
	c.Assert(orderCreate2.Side(), check.Equals, orderCreate1.Side())
	c.Assert(orderCreate2.Units(), check.Equals, orderCreate1.Units())
	c.Assert(orderCreate2.Price(), check.Equals, orderCreate1.Price())
	c.Assert(orderCreate2.Expiry(), check.Equals, orderCreate1.Expiry())
	c.Assert(orderCreate2.Reason(), check.Equals, orderCreate1.Reason())
}

func (ts *TestEventSuite) TestEventServer(c *check.C) {
	es, err := ts.c.NewEventServer(ts.c.AccountId())
	c.Assert(err, check.IsNil)

	wg := sync.WaitGroup{}

	t := time.AfterFunc(5*time.Minute, func() {
		es.Stop()
		c.Fail()
	})

	expiry := time.Now().Add(24 * time.Hour)

	wg.Add(1)
	go func() {
		err := es.ConnectAndHandle(func(accountId int, evt oanda.Event) {
			c.Log(accountId, evt)

			es.Stop()
			t.Stop()

			c.Assert(accountId, check.Equals, evt.AccountId())
			c.Assert(evt.Type(), check.Equals, "LIMIT_ORDER_CREATE")

			orderCreate, ok := evt.(*oanda.OrderCreateEvent)
			c.Assert(ok, check.Equals, true)
			c.Assert(orderCreate.Instrument(), check.Equals, "EUR_USD")
			c.Assert(orderCreate.Side(), check.Equals, "buy")
			c.Assert(orderCreate.Units(), check.Equals, 1)
			c.Assert(orderCreate.Price(), check.Equals, 0.75)

			orderExpiry := orderCreate.Expiry().Time()
			c.Assert(orderExpiry.Equal(expiry.Truncate(time.Second)), check.Equals, true)
			c.Assert(orderCreate.Reason(), check.Equals, "CLIENT_REQUEST")
		})
		c.Assert(err, check.IsNil)
		wg.Done()
	}()

	time.Sleep(5 * time.Second)

	ts.c.NewOrder(oanda.Limit, oanda.Buy, 1, "eur_usd", 0.75, expiry)
	wg.Wait()
}
