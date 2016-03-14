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
	"gopkg.in/check.v1"
)

type TestAccountSuite struct {
	OandaSuite
}

var _ = check.Suite(&TestAccountSuite{})

func (ts *TestAccountSuite) TestAccountApi(c *check.C) {
	accs, err := ts.Client.Accounts()
	c.Assert(err, check.IsNil)
	c.Logf("Accounts (%d): %v", len(accs), accs)
	c.Assert(len(accs) > 0, check.Equals, true)

	var idx int
	for i, acc := range accs {
		if acc.Name == "Primary" {
			idx = i
			break
		}
	}

	c.Assert(idx >= 0 && idx < len(accs), check.Equals, true)

	acc, err := ts.Client.Account(accs[idx].AccountId)
	c.Assert(err, check.IsNil)
	c.Log("Account:", acc)
	c.Assert(acc.AccountId, check.Not(check.Equals), 0)
	c.Assert(acc.Name, check.Equals, accs[idx].Name)
	c.Assert(acc.Currency, check.Equals, accs[idx].Currency)
	c.Assert(acc.Balance > 0, check.Equals, true)
	c.Assert(acc.MarginAvailable > 0, check.Equals, true)
	c.Assert(acc.MarginRate > 0, check.Equals, true)
	c.Assert(acc.MarginUsed, check.Equals, 0.0)
}
