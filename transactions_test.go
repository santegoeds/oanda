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

	"gopkg.in/check.v1"
)

type TestTranSuite struct {
	c *oanda.Client
}

var _ = check.Suite(&TestTranSuite{})

func (ts *TestTranSuite) SetUpSuite(c *check.C) {
	var err error
	ts.c, err = newSandboxClientWithAccount()
	c.Assert(err, check.IsNil)
}

func (ts *TestTranSuite) TestTransactionApi(c *check.C) {
	trans, err := ts.c.Transactions()
	c.Assert(err, check.IsNil)
	c.Assert(trans, check.HasLen, 2)

	m := make(map[string]bool)
	for _, tran := range trans {
		m[tran.Type()] = true

		switch tran.Type() {
		case "CREATE":
			_, err = tran.AsAccountCreate()
			c.Assert(err, check.IsNil)
		case "TRANSFER_FUNDS":
			_, err = tran.AsTransferFunds()
			c.Assert(err, check.IsNil)
		}
	}

	c.Log(m)

	_, ok := m["CREATE"]
	c.Assert(ok, check.Equals, true)

	_, ok = m["TRANSFER_FUNDS"]
	c.Assert(ok, check.Equals, true)

	tran, err := ts.c.Transaction(trans[0].TranId())
	c.Assert(err, check.IsNil)
	c.Log(tran)
	c.Assert(tran.Type(), check.Equals, "TRANSFER_FUNDS")

	tfTran, err := tran.AsTransferFunds()
	c.Assert(err, check.IsNil)
	c.Assert(tfTran.Amount(), check.Equals, 100000.0)
}
