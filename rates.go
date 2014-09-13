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
	"strings"
	"time"
)

type InstrumentInfo struct {
	DisplayName     string  `json:"displayName"`
	Pip             float64 `json:"pip,string"`
	MaxTradeUnits   int     `json:"maxTradeUnits"`
	Precision       float64 `json:"precision"`
	MaxTrailingStop float64 `json:"maxTrailingStop"`
	MinTrailingStop float64 `json:"minTrailingStop"`
	MarginRate      float64 `json:"marginRate"`
	Halted          bool    `json:"halted"`
	InterestRate    map[string]struct {
		Bid float64 `json:"bid"`
		Ask float64 `json:"ask"`
	} `json:"interestRate"`
}

func (ii InstrumentInfo) String() string {
	return fmt.Sprintf("InstrumentInfo{DisplayName: %s, Pip: %s, MarginRate: %f}", ii.DisplayName,
		ii.Pip, ii.MarginRate)
}

type InstrumentField string

const (
	If_DisplayName     InstrumentField = "displayName"
	If_Pip             InstrumentField = "pip"
	If_MaxTradeUnits   InstrumentField = "maxTradeUnits"
	If_Precision       InstrumentField = "precision"
	If_MaxTrailingStop InstrumentField = "maxTrailingStop"
	If_MinTrailingStop InstrumentField = "minTrailingStop"
	If_MarginRate      InstrumentField = "marginRate"
	If_Halted          InstrumentField = "halted"
	If_InterestRate    InstrumentField = "interestRate"
)

// Instruments returns the information of all instruments known to Oanda.
func (c *Client) Instruments(instruments []string, fields []InstrumentField) (
	map[string]InstrumentInfo, error) {

	u := c.getUrl("/v1/instruments", "api")
	q := u.Query()
	if len(instruments) > 0 {
		q.Set("instruments", strings.Join(instruments, ","))
	}
	if len(fields) > 0 {
		ss := make([]string, len(fields))
		for i, v := range fields {
			ss[i] = string(v)
		}
		q.Set("fields", strings.Join(ss, ","))
	}
	u.RawQuery = q.Encode()
	ctx, err := c.newContext("GET", u, nil)
	if err != nil {
		return nil, err
	}

	v := struct {
		ApiError
		Instruments []struct {
			Instrument string `json:"instrument"`
			InstrumentInfo
		} `json:"instruments"`
	}{}
	if _, err = ctx.Decode(&v); err != nil {
		return nil, err
	}

	info := make(map[string]InstrumentInfo)
	for _, in := range v.Instruments {
		info[in.Instrument] = in.InstrumentInfo
	}

	return info, nil
}

type (
	Granularity string
)

const (
	S5  Granularity = "S5"
	S10 Granularity = "S10"
	S15 Granularity = "S15"
	S30 Granularity = "S30"
	M1  Granularity = "M1"
	M2  Granularity = "M2"
	M3  Granularity = "M3"
	M5  Granularity = "M5"
	M10 Granularity = "M10"
	M15 Granularity = "M15"
	M30 Granularity = "M30"
	H1  Granularity = "H1"
	H2  Granularity = "H2"
	H3  Granularity = "H3"
	H4  Granularity = "H4"
	H6  Granularity = "H6"
	H8  Granularity = "H8"
	H12 Granularity = "H12"
	D   Granularity = "D"
	W   Granularity = "W"
	M   Granularity = "M"
)

type CandlesArg interface {
	ApplyCandlesArg(url.Values)
}

type (
	StartTime         time.Time
	EndTime           time.Time
	IncludeFirst      bool
	DailyAlignment    int
	AlignmentTimezone time.Location
	WeeklyAlignment   time.Weekday
)

func (c Count) ApplyCandlesArg(v url.Values) {
	optionalArgs(v).SetInt("count", int(c))
}

func (s StartTime) ApplyCandlesArg(v url.Values) {
	optionalArgs(v).SetTime("start", time.Time(s))
}

func (e EndTime) ApplyCandlesArg(v url.Values) {
	optionalArgs(v).SetTime("end", time.Time(e))
}

func (b IncludeFirst) ApplyCandlesArg(v url.Values) {
	optionalArgs(v).SetBool("includeFirst", bool(b))
}

func (da DailyAlignment) ApplyCandlesArg(v url.Values) {
	optionalArgs(v).SetInt("dailyAlignment", int(da))
}

func (atz AlignmentTimezone) ApplyCandlesArg(v url.Values) {
	loc := time.Location(atz)
	v.Set("alignmentTimezone", loc.String())
}

func (wa WeeklyAlignment) ApplyCandlesArg(v url.Values) {
	optionalArgs(v).SetStringer("weeklyAlignment", time.Weekday(wa))
}

type MidpointCandles struct {
	Instrument  string      `json:"instrument"`
	Granularity Granularity `json:"granularity"`
	Candles     []struct {
		Time     time.Time `json:"time"`
		OpenMid  float64   `json:"openMid"`
		HighMid  float64   `json:"highMid"`
		LowMid   float64   `json:"lowMid"`
		CloseMid float64   `json:"closeMid"`
		Volume   int       `json:"volume"`
		Complete bool      `json:"complete"`
	} `json:"candles"`
}

type BidAskCandles struct {
	Instrument  string      `json:"instrument"`
	Granularity Granularity `json:"granularity"`
	Candles     []struct {
		Time     time.Time `json:"time"`
		OpenBid  float64   `json:"openBid"`
		OpenAsk  float64   `json:"openAsk"`
		HighBid  float64   `json:"highBid"`
		HighAsk  float64   `json:"highAsk"`
		LowBid   float64   `json:"lowBid"`
		LowAsk   float64   `json:"lowAsk"`
		CloseBid float64   `json:"closeBid"`
		CloseAsk float64   `json:"closeAsk"`
		Volume   int       `json:"volume"`
		Complete bool      `json:"complete"`
	} `json:"candles"`
}

// MidpointCandles returns historic price information for an instrument.
func (c *Client) MidpointCandles(instrument string, granularity Granularity,
	args ...CandlesArg) (*MidpointCandles, error) {

	ctx, err := c.newCandlesContext(instrument, granularity, "midpoint", args...)
	if err != nil {
		return nil, err
	}

	candles := struct {
		ApiError
		MidpointCandles
	}{}
	if _, err = ctx.Decode(&candles); err != nil {
		return nil, err
	}

	return &candles.MidpointCandles, nil
}

// BidAskCandles returns historic price information for an instrument.
func (c *Client) BidAskCandles(instrument string, granularity Granularity,
	args ...CandlesArg) (*BidAskCandles, error) {

	ctx, err := c.newCandlesContext(instrument, granularity, "bidask", args...)
	if err != nil {
		return nil, err
	}

	candles := struct {
		ApiError
		BidAskCandles
	}{}
	if _, err = ctx.Decode(&candles); err != nil {
		return nil, err
	}

	return &candles.BidAskCandles, nil
}

func (c *Client) newCandlesContext(instrument string, granularity Granularity, candleFormat string,
	args ...CandlesArg) (*Context, error) {

	u := c.getUrl("/v1/candles", "api")
	q := u.Query()
	q.Set("candleFormat", candleFormat)
	for _, arg := range args {
		arg.ApplyCandlesArg(q)
	}
	u.RawQuery = q.Encode()

	return c.newContext("GET", u, nil)
}
