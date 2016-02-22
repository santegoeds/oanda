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

func (oa optionalArgs) SetInt64(k string, n int64) {
	url.Values(oa).Set(k, strconv.FormatInt(n, 10))
}

func (oa optionalArgs) SetId(k string, id Id) {
	url.Values(oa).Set(k, strconv.FormatUint(uint64(id), 10))
}

func (oa optionalArgs) SetFloat(k string, f float64) {
	url.Values(oa).Set(k, strconv.FormatFloat(f, 'f', -1, 64))
}

func (oa optionalArgs) SetIdArray(k string, ia []Id) {
	switch n := len(ia); {
	case n == 0:
		return
	case n == 1:
		url.Values(oa).Set(k, strconv.FormatUint(uint64(ia[0]), 10))
	default:
		strIds := make([]string, n)
		for i, v := range ia {
			strIds[i] = strconv.FormatUint(uint64(v), 10)
		}
		url.Values(oa).Set("ids", strings.Join(strIds, ","))
	}
}

func (oa optionalArgs) SetTime(k string, t time.Time) {
	oa.SetInt64(k, t.UTC().Unix())
}

func (oa optionalArgs) SetStringer(k string, v fmt.Stringer) {
	url.Values(oa).Set(k, v.String())
}

func (oa optionalArgs) SetBool(k string, b bool) {
	url.Values(oa).Set(k, strconv.FormatBool(b))
}

type Time string

// Time return the time as time.Time instance.
func (t Time) Time() time.Time {
	return time.Unix(0, t.UnixNano())
}

func (t Time) UnixMicro() int64 {
	if t.IsZero() {
		return 0
	}
	if unix, err := strconv.ParseInt(string(t), 10, 64); err == nil {
		return unix
	}
	return 0
}

func (t Time) UnixNano() int64 {
	return t.UnixMicro() * 1000
}

func (t Time) String() string {
	if !t.IsZero() {
		return t.Time().Format(time.RFC3339)
	}
	return string(t)
}

func (t Time) IsZero() bool {
	return t == ""
}
