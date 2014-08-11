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
	"strings"

	"oanda"

	"gopkg.in/check.v1"
)

type DecodeSuite struct{}

var _ = check.Suite(&DecodeSuite{})

const (
	okData = `
        {
            "A": "A", 
            "B": 1, 
            "C": 2, 
            "x": "X",
            "Nested": {
                "E": 3
            }

        }
        `
	errorData = `
        {
            "A": "A", 
            "B": 1, 
            "C": 1, 
            "x": "X", 
            "code": 1, 
            "message": "Test Message",
            "moreinfo": "More Info"
        }
        `
)

type DecodeStruct struct {
	A      string
	B      float64
	C      int
	D      string `json:"x"`
	Nested struct {
		E int
	}
}

// TestDecodeStruct verifies that a Json byte string can be Decoded into a struct.
func (s *DecodeSuite) TestDecodeStruct(c *check.C) {
	r := strings.NewReader(okData)
	dec := oanda.NewDecoder(r)

	v := DecodeStruct{}
	if err := dec.Decode(&v); err != nil {
		c.Error(err)
		return
	}

	c.Assert(v.A, check.Equals, "A")
	c.Assert(v.B, check.Equals, 1.0)
	c.Assert(v.C, check.Equals, 2)
	c.Assert(v.D, check.Equals, "X")
	c.Assert(v.Nested.E, check.Equals, 3)
}

// TestDecodeMap verifies that a Json byte string can be Decoded into a map[string]interface{}
func (s *DecodeSuite) TestDecodeMap(c *check.C) {
	r := strings.NewReader(okData)
	dec := oanda.NewDecoder(r)

	m := make(map[string]interface{}, 0)
	if err := dec.Decode(&m); err != nil {
		c.Error(err)
		return
	}

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

// TestDecodeStructWithApiError verifies that decoding a json string that includes Oanda error
// returns an ApiError if the json string is decoded into a struct.
func (s *DecodeSuite) TestDecodeStructWithApiError(c *check.C) {
	r := strings.NewReader(errorData)
	dec := oanda.NewDecoder(r)

	v := DecodeStruct{}
	err := dec.Decode(&v)
	if err == nil {
		c.Error("Expected Decode to return oanda.ApiError")
		return
	}

	apiError, ok := err.(*oanda.ApiError)
	if !ok {
		c.Error("Expected error as ApiError")
		return
	}

	c.Assert(apiError.Code, check.Equals, 1)
	c.Assert(apiError.Message, check.Equals, "Test Message")
	c.Assert(apiError.MoreInfo, check.Equals, "More Info")
}

// TestDecodeMapWithApiError verifies that decoding a json string that includes an Oanda error
// returns an ApiError if the json string is decoded into a map[string]interface{}
func (s *DecodeSuite) TestDecodeMapWithApiError(c *check.C) {
	r := strings.NewReader(errorData)
	dec := oanda.NewDecoder(r)

	m := make(map[string]interface{})
	err := dec.Decode(&m)
	if err == nil {
		c.Error("Expected Decode to return oanda.ApiError")
		return
	}

	apiError, ok := err.(*oanda.ApiError)
	if !ok {
		c.Error("Expected error as ApiError")
		return
	}

	c.Assert(apiError.Code, check.Equals, 1)
	c.Assert(apiError.Message, check.Equals, "Test Message")
	c.Assert(apiError.MoreInfo, check.Equals, "More Info")
}
