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

package oanda

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type optionalArgs url.Values

func (oa optionalArgs) SetInt(k string, n int) {
	url.Values(oa).Set(k, strconv.Itoa(n))
}

func (oa optionalArgs) SetFloat(k string, f float64) {
	url.Values(oa).Set(k, strconv.FormatFloat(f, 'f', -1, 64))
}

func (oa optionalArgs) SetIntArray(k string, ia []int) {
	switch n := len(ia); {
	case n == 0:
		return
	case n == 1:
		url.Values(oa).Set(k, strconv.Itoa(ia[0]))
	default:
		strIds := make([]string, n)
		for i, v := range ia {
			strIds[i] = strconv.Itoa(v)
		}
		url.Values(oa).Set("ids", strings.Join(strIds, ","))
	}
}

func (oa optionalArgs) SetTime(k string, t time.Time) {
	url.Values(oa).Set(k, t.UTC().Format(time.RFC3339))
}

func (oa optionalArgs) SetStringer(k string, v fmt.Stringer) {
	url.Values(oa).Set(k, v.String())
}

func (oa optionalArgs) SetBool(k string, b bool) {
	url.Values(oa).Set(k, strconv.FormatBool(b))
}

// Time embeds time.Time and serves to Unmarshal from Integer values.
type Time struct {
	time.Time
}

// Implements the json.UnmarshalJSON interface.
func (t *Time) UnmarshalJSON(data []byte) error {
	var secs int64
	if err := json.Unmarshal(data, &secs); err != nil {
		return err
	}
	t.Time = time.Unix(secs, 0)
	return nil
}
