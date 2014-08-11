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
	"testing"

	"oanda"

	"gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type TestSuite struct {
	c *oanda.Client
}

var _ = check.Suite(&TestSuite{})

func (ts *TestSuite) SetUpSuite(c *check.C) {
	sbClient := oanda.NewSandboxClient()
	accountId, err := sbClient.NewAccount()
	if err != nil {
		c.Error(err)
		return
	}
	ts.c = sbClient.Client
	ts.c.AccountId = accountId
	return
}

func (ts *TestSuite) TestAccount(c *check.C) {
	acc, err := ts.c.Account(ts.c.AccountId)
	if err != nil {
		c.Error(err)
		return
	}

	c.Log("Account:", acc)

	c.Assert(acc.AccountId, check.Not(check.Equals), 0)
	c.Assert(acc.Name, check.Equals, "Primary")
	c.Assert(acc.Currency, check.Equals, "USD")
}

func (ts *TestSuite) TestAccounts(c *check.C) {
	accs, err := ts.c.Accounts()
	if err != nil {
		c.Error(err)
		return
	}

	c.Logf("Accounts (%d): %v", len(accs), accs)

	c.Assert(accs, check.HasLen, 1)
	c.Assert(accs[0].Name, check.Equals, "Primary")
	c.Assert(accs[0].Currency, check.Equals, "USD")
}
