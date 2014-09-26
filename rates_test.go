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

func (ts *TestSuite) TestInstruments(c *check.C) {
	instruments, err := ts.c.Instruments(nil, nil)
	c.Assert(err, check.IsNil)
	c.Log(instruments)
	c.Assert(instruments, check.Not(check.HasLen), 0)
}
