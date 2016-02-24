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

type InterestRate struct {
	Bid float64 `json:"bid"`
	Ask float64 `json:"ask"`
}

func (ir InterestRate) String() string {
	return fmt.Sprintf("InterestRate{Bid: %v, Ask: %v}", ir.Bid, ir.Ask)
}

type InstrumentInfo struct {
	DisplayName     string                  `json:"displayName"`
	Pip             float64                 `json:"pip,string"`
	MaxTradeUnits   int                     `json:"maxTradeUnits"`
	Precision       float64                 `json:"precision,string"`
	MaxTrailingStop float64                 `json:"maxTrailingStop"`
	MinTrailingStop float64                 `json:"minTrailingStop"`
	MarginRate      float64                 `json:"marginRate"`
	Halted          bool                    `json:"halted"`
	InterestRate    map[string]InterestRate `json:"interestRate"`
}

func (ii InstrumentInfo) String() string {
	return fmt.Sprintf(
		"InstrumentInfo{\n"+
			"    DisplayName: %v,\n"+
			"    Pip: %v,\n"+
			"    MaxTradeUnits: %v\n"+
			"    Precision: %v\n"+
			"    MaxTrailingStop: %v\n"+
			"    MinTrailingStop: %v\n"+
			"    MarginRate: %v\n"+
			"    Halted: %v\n"+
			"    InterestRate: %v\n"+
			"}",
		ii.DisplayName, ii.Pip, ii.MaxTradeUnits, ii.Precision, ii.MaxTrailingStop,
		ii.MinTrailingStop, ii.MarginRate, ii.Halted, ii.InterestRate)
}

type InstrumentField string

const (
	DisplayNameField     InstrumentField = "displayName"
	PipField             InstrumentField = "pip"
	MaxTradeUnitsField   InstrumentField = "maxTradeUnits"
	PrecisionField       InstrumentField = "precision"
	MaxTrailingStopField InstrumentField = "maxTrailingStop"
	MinTrailingStopField InstrumentField = "minTrailingStop"
	MarginRateField      InstrumentField = "marginRate"
	HaltedField          InstrumentField = "halted"
	InterestRateField    InstrumentField = "interestRate"
)

// Instruments returns instrument information.  Only the specified instruments are returned if instruments
// is not nil.  If fields is not nil additional information fields is included.
//
// See http://developer.oanda.com/docs/v1/rates/#get-an-instrument-list for further information.
func (c *Client) Instruments(instruments []string, fields []InstrumentField) (map[string]InstrumentInfo, error) {

	u, err := url.Parse("/v1/instruments")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	if len(instruments) > 0 {
		q.Set("instruments", strings.ToUpper(strings.Join(instruments, ",")))
	}
	if len(fields) > 0 {
		ss := make([]string, len(fields))
		for i, v := range fields {
			ss[i] = string(v)
		}
		q.Set("fields", strings.Join(ss, ","))
	}
	if c.accountId != 0 {
		q.Set("accountId", strconv.FormatUint(uint64(c.accountId), 10))
	}
	u.RawQuery = q.Encode()

	v := struct {
		Instruments []struct {
			Instrument string `json:"instrument"`
			InstrumentInfo
		} `json:"instruments"`
	}{}
	if err = getAndDecode(c, u.String(), &v); err != nil {
		return nil, err
	}

	info := make(map[string]InstrumentInfo)
	for _, in := range v.Instruments {
		info[in.Instrument] = in.InstrumentInfo
	}

	return info, nil
}

type (
	// Granularity determines the interval at which historic instrument prices are converted into candles.
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

// CandlesArg implements optional arguments for MidpointCandles and BidAskCandles.
type CandlesArg interface {
	applyCandlesArg(url.Values)
}

type (
	// Optional argument for PollMidpriceCandles and PollBidAskCandles to specify the start time from which
	// instrument history should be included.
	//
	// See http://developer.oanda.com/docs/v1/rates/#retrieve-instrument-history for further information.
	StartTime time.Time

	// Optional argument for PollMidpriceCandles and PollBidAskCandles to specify the end time of until which
	// instrument history should be included.
	//
	// See http://developer.oanda.com/docs/v1/rates/#retrieve-instrument-history for further information.
	EndTime time.Time

	// Optional argument for PollMidpriceCandles and PollBidAskCandles to indicate whether the candle that
	// starts at StartTime should be included.
	//
	// See http://developer.oanda.com/docs/v1/rates/#retrieve-instrument-history for further information.
	IncludeFirst bool

	// Optional argument for PollMidpriceCandles and PollBidAskCandles to indicate the hour at which
	// candles hould be aligned.  Only relevant for hourly or greater granularities.
	//
	// See http://developer.oanda.com/docs/v1/rates/#retrieve-instrument-history for further information.
	DailyAlignment int

	// Optional argument for PollMidpriceCandles and PollBidAskCandles that indicates the timezone to use
	// when aligning candles with DailyAlignment.
	//
	// See http://developer.oanda.com/docs/v1/rates/#retrieve-instrument-history for further information.
	AlignmentTimezone time.Location

	// Optional argument for PollMidpriceCandles and PollBidAskCandles to indicate the weekday at which
	// candles should be aligned. Only relevant for weekly granularity.
	//
	// See http://developer.oanda.com/docs/v1/rates/#retrieve-instrument-history for further information.
	WeeklyAlignment time.Weekday
)

func (c Count) applyCandlesArg(v url.Values) {
	optionalArgs(v).SetInt("count", int(c))
}

func (s StartTime) applyCandlesArg(v url.Values) {
	optionalArgs(v).SetTime("start", time.Time(s))
}

func (e EndTime) applyCandlesArg(v url.Values) {
	optionalArgs(v).SetTime("end", time.Time(e))
}

func (b IncludeFirst) applyCandlesArg(v url.Values) {
	optionalArgs(v).SetBool("includeFirst", bool(b))
}

func (da DailyAlignment) applyCandlesArg(v url.Values) {
	optionalArgs(v).SetInt("dailyAlignment", int(da))
}

func (atz AlignmentTimezone) applyCandlesArg(v url.Values) {
	loc := time.Location(atz)
	v.Set("alignmentTimezone", loc.String())
}

func (wa WeeklyAlignment) applyCandlesArg(v url.Values) {
	optionalArgs(v).SetStringer("weeklyAlignment", time.Weekday(wa))
}

// MidpointCandles represents instrument history with a specific granularity.
type MidpointCandles struct {
	Instrument  string           `json:"instrument"`
	Granularity Granularity      `json:"granularity"`
	Candles     []MidpointCandle `json:"candles"`
}

func (c MidpointCandles) String() string {
	return fmt.Sprintf("MidpointCandles{Instrument: %s, Granularity: %v, Candles: %v}",
		c.Instrument, c.Granularity, c.Candles)
}

// BidAskCandles represents Bid and Ask instrument history with a specific granularity.
type BidAskCandles struct {
	Instrument  string         `json:"instrument"`
	Granularity Granularity    `json:"granularity"`
	Candles     []BidAskCandle `json:"candles"`
}

func (c BidAskCandles) String() string {
	return fmt.Sprintf("BidAskCandles{Instrument: %s, Granularity: %v, Candles: %v}", c.Instrument,
		c.Granularity, c.Candles)
}

// PollMidpointCandles returns historical midpoint prices for an instrument.
func (c *Client) PollMidpointCandles(instrument string, granularity Granularity,
	args ...CandlesArg) (*MidpointCandles, error) {

	u, err := c.newCandlesURL(instrument, granularity, "midpoint", args...)
	if err != nil {
		return nil, err
	}
	candles := MidpointCandles{}
	if err = getAndDecode(c, u.String(), &candles); err != nil {
		return nil, err
	}
	return &candles, nil
}

// PollBidAskCandles returns historical bid- and ask prices for an instrument.
func (c *Client) PollBidAskCandles(instrument string, granularity Granularity,
	args ...CandlesArg) (*BidAskCandles, error) {

	u, err := c.newCandlesURL(instrument, granularity, "bidask", args...)
	if err != nil {
		return nil, err
	}
	candles := BidAskCandles{}
	if err = getAndDecode(c, u.String(), &candles); err != nil {
		return nil, err
	}
	return &candles, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Private

func (c *Client) newCandlesURL(instrument string, granularity Granularity, candleFormat string,
	args ...CandlesArg) (*url.URL, error) {

	u, err := url.Parse("/v1/candles")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("candleFormat", candleFormat)
	q.Set("granularity", string(granularity))
	q.Set("instrument", strings.ToUpper(instrument))
	for _, arg := range args {
		arg.applyCandlesArg(q)
	}
	u.RawQuery = q.Encode()

	return u, err
}

type MidpointCandle struct {
	Time     Time    `json:"time"`
	OpenMid  float64 `json:"openMid"`
	HighMid  float64 `json:"highMid"`
	LowMid   float64 `json:"lowMid"`
	CloseMid float64 `json:"closeMid"`
	Volume   int     `json:"volume"`
	Complete bool    `json:"complete"`
}

func (c MidpointCandle) String() string {
	return fmt.Sprintf("MidpointCandle{Time: %v, OpenMid: %f, HighMid: %f, LowMid: %f, "+
		"CloseMid: %f, Volume: %d, Complete: %v}", c.Time, c.OpenMid, c.HighMid, c.LowMid,
		c.CloseMid, c.Volume, c.Complete)
}

type BidAskCandle struct {
	Time     Time    `json:"time"`
	OpenBid  float64 `json:"openBid"`
	OpenAsk  float64 `json:"openAsk"`
	HighBid  float64 `json:"highBid"`
	HighAsk  float64 `json:"highAsk"`
	LowBid   float64 `json:"lowBid"`
	LowAsk   float64 `json:"lowAsk"`
	CloseBid float64 `json:"closeBid"`
	CloseAsk float64 `json:"closeAsk"`
	Volume   int     `json:"volume"`
	Complete bool    `json:"complete"`
}

func (c BidAskCandle) String() string {
	return fmt.Sprintf("BidAskCandle{Time: %v, OpenBid: %f, OpenAsk: %f, HighBid: %f, "+
		"HighAsk: %f, LowBid: %f, LowAsk: %f, CloseBid: %f, CloseAsk: %f, "+
		"Volume: %d, Complete: %v}", c.Time, c.OpenBid, c.OpenAsk, c.HighBid,
		c.HighAsk, c.LowBid, c.LowAsk, c.CloseBid, c.CloseAsk, c.Volume, c.Complete)
}
