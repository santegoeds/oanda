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
	"encoding/json"
	"strings"

	"github.com/santegoeds/oanda"

	"gopkg.in/check.v1"
)

type TestDecodeSuite struct{}

var _ = check.Suite(&TestDecodeSuite{})

const (
	okData = `{
            "A": "A", 
            "B": 1, 
            "C": 2, 
            "x": "X",
            "Nested": { "E": 3 }
        }`
	errorData = `{
            "A": "A", 
            "B": 1, 
            "C": 1, 
            "x": "X", 
            "code": 1, 
            "message": "Test Message",
            "moreInfo": "More Info"
        }`
	partialErrorData = `{
            "A": "A",
            "B": 1,
            "C": 1,
            "x": "X",
            "code": 1
        }`
)

type StructReceiver struct {
	A      string
	B      float64
	C      int
	D      string `json:"x"`
	Nested struct {
		E int
	}
}

type StructReceiverWithApiError struct {
	StructReceiver
	oanda.ApiError
}

func (s *TestDecodeSuite) TestDecodeStructReceiver(c *check.C) {
	dec := oanda.NewDecoder(strings.NewReader(okData))

	srw := StructReceiverWithApiError{}
	err := dec.Decode(&srw)
	c.Assert(err, check.IsNil)

	c.Assert(srw.A, check.Equals, "A")
	c.Assert(srw.B, check.Equals, 1.0)
	c.Assert(srw.C, check.Equals, 2)
	c.Assert(srw.D, check.Equals, "X")
	c.Assert(srw.Nested.E, check.Equals, 3)

	// Verify that an error is returned if the receiver struct does not have a Code field.
	dec = oanda.NewDecoder(strings.NewReader(okData))
	sr := StructReceiver{}
	err = dec.Decode(&sr)
	c.Assert(err, check.NotNil)

	_, ok := err.(*oanda.ApiError)
	c.Assert(ok, check.Equals, false)
}

// TestDecodeMap verifies that a Json byte string can be Decoded into a map[string]interface{}
func (s *TestDecodeSuite) TestDecodeInterfaceMap(c *check.C) {
	dec := oanda.NewDecoder(strings.NewReader(okData))
	m := make(map[string]interface{}, 0)
	err := dec.Decode(&m)
	c.Assert(err, check.IsNil)
	c.Log(m)

	c.Assert(m["A"], check.Equals, "A")
	c.Assert(m["B"], check.Equals, 1.0)
	c.Assert(m["C"], check.Equals, 2.0)
	c.Assert(m["x"], check.Equals, "X")

	nm, ok := m["Nested"].(map[string]interface{})
	if !ok {
		c.Error("Nested map is of an unexpected type")
		return
	}
	c.Assert(nm["E"], check.Equals, 3.0)
}

func (s *TestDecodeSuite) TestDecodeRawMessageMap(c *check.C) {
	dec := oanda.NewDecoder(strings.NewReader(okData))
	m := make(map[string]json.RawMessage)
	err := dec.Decode(&m)
	c.Assert(err, check.IsNil)

	c.Log(m)

	c.Assert(string(m["A"]), check.Equals, `"A"`)
	c.Assert(string(m["B"]), check.Equals, "1")
	c.Assert(string(m["C"]), check.Equals, "2")
	c.Assert(string(m["x"]), check.Equals, `"X"`)
	c.Assert(string(m["Nested"]), check.Equals, `{ "E": 3 }`)
}

func (s *TestDecodeSuite) TestDecodeApiErrorFromJson(c *check.C) {
	sr := StructReceiverWithApiError{}
	testDecodeErrorFromJson(&sr, c)

	im := make(map[string]interface{})
	testDecodeErrorFromJson(&im, c)

	rm := make(map[string]json.RawMessage)
	testDecodeErrorFromJson(&rm, c)
}

func testDecodeErrorFromJson(vp interface{}, c *check.C) {
	r := strings.NewReader(errorData)
	dec := oanda.NewDecoder(r)
	err := dec.Decode(vp)

	c.Assert(err, check.NotNil)

	apiErr, ok := err.(*oanda.ApiError)
	c.Assert(ok, check.Equals, true)
	c.Assert(apiErr.Code, check.Equals, 1)
	c.Assert(apiErr.Message, check.Equals, "Test Message")
	c.Assert(apiErr.MoreInfo, check.Equals, "More Info")
}
