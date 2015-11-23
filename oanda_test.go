/*
   Copyright 2014 Tjerk Santegoeds

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package oanda_test

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"gopkg.in/check.v1"

	"github.com/santegoeds/oanda"
)

type OandaSuite struct {
	Client      *oanda.Client
	OpenMarkets []string
}

func Test(t *testing.T) { check.TestingT(t) }

func (s *OandaSuite) SetUpSuite(c *check.C) {
	envName := "FXPRACTICE_TOKEN"
	token := os.Getenv(envName)
	if token == "" {
		c.Skip(fmt.Sprintf("Environment variable %s is not defined", envName))
		return
	}
	time.Sleep(2 * time.Second)
	client, err := oanda.NewFxPracticeClient(token)

	c.Assert(err, check.IsNil)
	c.Assert(client, check.NotNil)

	s.Client = client
}

func (s *OandaSuite) SetUpAccount(c *check.C) {
	envName := "FXPRACTICE_ACCOUNT"
	accountIdStr := os.Getenv(envName)
	if accountIdStr == "" {
		c.Skip(fmt.Sprintf("Environment variable %s is not defined", envName))
		return
	}

	accountId, err := strconv.Atoi(accountIdStr)
	c.Assert(err, check.IsNil)
	s.Client.SelectAccount(accountId)

	CancelAllOrders(c, s.Client)
	CloseAllPositions(c, s.Client)
}

func (s *OandaSuite) TearDownSuite(c *check.C) {
	if s.Client != nil && s.Client.AccountId() != 0 {
		CancelAllOrders(c, s.Client)
		CloseAllPositions(c, s.Client)
	}
}

func CancelAllOrders(c *check.C, client *oanda.Client) {
	if client == nil {
		return
	}

	orders, err := client.Orders()
	c.Assert(err, check.IsNil)

	for _, o := range orders {
		_, err := client.CancelOrder(o.OrderId)
		c.Assert(err, check.IsNil)
	}
}

func CloseAllPositions(c *check.C, client *oanda.Client) {
	if client == nil {
		return
	}

	positions, err := client.Positions()
	c.Assert(err, check.IsNil)

	for _, pos := range positions {
		_, err = client.ClosePosition(pos.Instrument)
		c.Assert(err, check.IsNil)
	}
}

type Counter struct {
	m sync.RWMutex
	n int
}

func (c *Counter) Inc() int {
	c.m.Lock()
	defer c.m.Unlock()
	c.n++
	return c.n
}

func (c *Counter) Val() int {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.n
}
