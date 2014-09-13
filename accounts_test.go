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

	"github.com/santegoeds/oanda"

	"gopkg.in/check.v1"
)

func newSandboxClientWithAccount() (*oanda.Client, error) {
	sbClient := oanda.NewSandboxClient()
	accountId, err := sbClient.NewAccount()
	if err != nil {
		return nil, err
	}
	c := sbClient.Client
	c.AccountId = accountId
	return c, nil
}

func Test(t *testing.T) { check.TestingT(t) }

type TestSuite struct {
	c *oanda.Client
}

var _ = check.Suite(&TestSuite{})

func (ts *TestSuite) SetUpSuite(c *check.C) {
	var err error
	ts.c, err = newSandboxClientWithAccount()
	c.Assert(err, check.IsNil)
}

func (ts *TestSuite) TestAccount(c *check.C) {
	acc, err := ts.c.Account(ts.c.AccountId)
	c.Assert(err, check.IsNil)
	c.Log("Account:", acc)
	c.Assert(acc.AccountId, check.Not(check.Equals), 0)
	c.Assert(acc.Name, check.Equals, "Primary")
	c.Assert(acc.Currency, check.Equals, "USD")
}

func (ts *TestSuite) TestAccounts(c *check.C) {
	accs, err := ts.c.Accounts()
	c.Assert(err, check.IsNil)
	c.Logf("Accounts (%d): %v", len(accs), accs)
	c.Assert(accs, check.HasLen, 1)
	c.Assert(accs[0].Name, check.Equals, "Primary")
	c.Assert(accs[0].Currency, check.Equals, "USD")
}
