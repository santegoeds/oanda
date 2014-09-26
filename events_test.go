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
	"sync"
	"time"

	"gopkg.in/check.v1"
)

type TestEventSuite struct {
	AccountId int
	c         *oanda.Client
}

var _ = check.Suite(&TestEventSuite{})

func (ts *TestEventSuite) SetUpTest(c *check.C) {
	client, err := oanda.NewSandboxClient()
	c.Assert(err, check.IsNil)
	ts.c = client

	accs, err := client.Accounts()
	c.Assert(err, check.IsNil)
	c.Assert(accs, check.HasLen, 1)

	ts.AccountId = accs[0].AccountId
	ts.c.SelectAccount(ts.AccountId)
}

func (ts *TestEventSuite) TestEventApi(c *check.C) {
	events, err := ts.c.PollEvents()
	c.Assert(err, check.IsNil)
	c.Assert(events, check.HasLen, 2)

	m := make(map[string]bool)
	for _, evt := range events {
		m[evt.Type()] = true

		switch evt.Type() {
		case "CREATE":
			accountCreate, ok := evt.(*oanda.AccountCreateEvent)
			c.Assert(ok, check.Equals, true)
			c.Check(accountCreate.HomeCurrency(), check.Not(check.Equals), "")
			c.Check(accountCreate.Reason(), check.Not(check.Equals), "")

		case "TRANSFER_FUNDS":
			transferFunds, ok := evt.(*oanda.TransferFundsEvent)
			c.Assert(ok, check.Equals, true)
			c.Check(transferFunds.Amount(), check.Equals, 100000.)

		}
	}

	c.Log(m)

	_, ok := m["CREATE"]
	c.Assert(ok, check.Equals, true)

	_, ok = m["TRANSFER_FUNDS"]
	c.Assert(ok, check.Equals, true)

	evt, err := ts.c.PollEvent(events[0].TranId())
	c.Assert(err, check.IsNil)

	c.Log(evt)
	c.Check(evt.Type(), check.Equals, events[0].Type())
	c.Check(evt.AccountId(), check.Equals, events[0].AccountId())
	c.Check(evt.Time(), check.Equals, events[0].Time())

	transferFunds1, ok := events[0].(*oanda.TransferFundsEvent)
	c.Assert(ok, check.Equals, true)

	transferFunds2, ok := evt.(*oanda.TransferFundsEvent)
	c.Assert(ok, check.Equals, true)

	c.Check(transferFunds1.Amount(), check.Equals, transferFunds2.Amount())
}

func (ts *TestEventSuite) TestEventServer(c *check.C) {
	es, err := ts.c.NewEventServer(ts.AccountId)
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
			c.Assert(orderCreate.Expiry().Equal(expiry.Truncate(time.Second)), check.Equals, true)
			c.Assert(orderCreate.Reason(), check.Equals, "CLIENT_REQUEST")
		})
		c.Assert(err, check.IsNil)
		wg.Done()
	}()

	time.Sleep(5 * time.Second)

	ts.c.NewOrder(oanda.Limit, oanda.Buy, 1, "eur_usd", 0.75, expiry)
	wg.Wait()
}
