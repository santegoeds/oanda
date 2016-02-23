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
	"sort"
	"strconv"
	"strings"
	"time"
)

type Period int64

const (
	Hour  Period = Period(time.Hour)
	Day   Period = 24 * Hour
	Week  Period = 7 * Day
	Month Period = 2592000
	Year  Period = 31536000
)

///////////////////////////////////////////////////////////////////////////////////////////////////
// Calendar

type CalendarEvent struct {
	Title     string  `json:"title"`
	Timestamp int64   `json:"timestamp"`
	Unit      string  `json:"unit"`
	Currency  string  `json:"currency"`
	Forecast  float64 `json:"forecast,string"`
	Previous  float64 `json:"previous,string"`
	Actual    float64 `json:"actual,string"`
	Market    float64 `json:"market,string"`
}

func (ce CalendarEvent) String() string {
	t := time.Unix(0, ce.Timestamp*1000)
	return fmt.Sprintf("CalendarEvent{Title: %s, Timestamp: %s, Unit: %s, Currency: %s, "+
		"Forecast: %v, Previous: %v, Actual: %v, Market: %v}", ce.Title,
		t.Format(time.RFC3339), ce.Unit, ce.Currency, ce.Forecast, ce.Previous, ce.Actual,
		ce.Market)
}

// Calendar returns and array of economic calendar events associated with an instrument. Events
// can include economic indicator data or they can solely be be news about important meetings.
//
// See http://developer.oanda.com/docs/v1/forex-labs/#calendar for further information.
func (c *Client) Calendar(instrument string, period Period) ([]CalendarEvent, error) {
	instrument = strings.ToUpper(instrument)
	u, err := url.Parse("/labs/v1/calendar")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("instrument", instrument)
	q.Set("period", strconv.Itoa(int(period)))
	u.RawQuery = q.Encode()

	ces := make([]CalendarEvent, 0)
	if err = getAndDecode(c, u.String(), &ces); err != nil {
		return nil, err
	}
	return ces, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// PositionRatios

type PositionRatio struct {
	Timestamp    Time
	LongRatio    float64
	ExchangeRate float64
}

type PositionRatios struct {
	Instrument  string
	DisplayName string
	Ratios      []PositionRatio
}

func (pr PositionRatios) String() string {
	return fmt.Sprintf("PositionRatios{Instrument: %s, DisplayName: %s, Ratios: %v}",
		pr.Instrument, pr.DisplayName, pr.Ratios)
}

func (pr *PositionRatios) UnmarshalJSON(data []byte) error {
	v := struct {
		Data map[string]struct {
			Data  [][]float64 `json:"data"`
			Label string      `json:"label"`
		} `json:"data"`
	}{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	for instrument, data := range v.Data {
		pr.Instrument = instrument
		pr.DisplayName = data.Label
		pr.Ratios = make([]PositionRatio, len(data.Data))
		for i, ratio := range data.Data {
			pr.Ratios[i].Timestamp = Time(strconv.Itoa(int(ratio[0])))
			pr.Ratios[i].LongRatio = ratio[1]
			pr.Ratios[i].ExchangeRate = ratio[2]
		}
	}
	return nil
}

// PositionRatios returns daily position ratios for an instrument. A position ratio is
// the percentage of Oanda clients that have a Long/Short position.
//
// See http://developer.oanda.com/docs/v1/forex-labs/#historical-position-ratios for further
// information.
func (c *Client) PositionRatios(instrument string, period Period) (*PositionRatios, error) {
	instrument = strings.ToUpper(instrument)
	u, err := url.Parse("/labs/v1/historical_position_ratios")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("instrument", instrument)
	q.Set("period", strconv.Itoa(int(period)))
	u.RawQuery = q.Encode()

	pr := PositionRatios{}
	if err = getAndDecode(c, u.String(), &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Spreads

type Spread struct {
	Timestamp Time
	Spread    float64
}

func (s Spread) String() string {
	return fmt.Sprintf("Spread{Timestamp: %v, Spread: %f}", s.Timestamp, s.Spread)
}

func (s *Spread) UnmarshalJSON(data []byte) error {
	v := []float64{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	s.Timestamp = Time(strconv.Itoa(int(v[0])))
	s.Spread = v[1]
	return nil
}

type Spreads struct {
	Max []Spread `json:"max"`
	Avg []Spread `json:"avg"`
	Min []Spread `json:"min"`
}

func (s Spreads) String() string {
	return fmt.Sprintf("Spreads{Max: %v, Avg: %v, Min: %v}", s.Max, s.Avg, s.Min)
}

// Spreads returns historical spread data for a specific period in 15 min intervals.  If unique is
// true then adjacent duplicate spreads are omitted.
//
// See http://developer.oanda.com/docs/v1/forex-labs/#spreads for further information.
func (c *Client) Spreads(instrument string, period Period, unique bool) (*Spreads, error) {
	instrument = strings.ToUpper(instrument)
	u, err := url.Parse("/labs/v1/spreads")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("instrument", instrument)
	q.Set("period", strconv.Itoa(int(period)))
	if unique {
		q.Set("unique", "1")
	} else {
		q.Set("unique", "0")
	}
	u.RawQuery = q.Encode()

	s := Spreads{}
	if err = getAndDecode(c, u.String(), &s); err != nil {
		return nil, err
	}
	return &s, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// CommitmentsOfTraders

type CommitmentsOfTraders struct {
	Date               int64   `json:"date"`
	Price              float64 `json:"price,string"`
	OverallInterest    int     `json:"oi,string"`
	NonCommercialLong  int     `json:"ncl,string"`
	NonCommercialShort int     `json:"ncs,string"`
	Unit               string  `json:"unit"`
}

func (c CommitmentsOfTraders) String() string {
	t := time.Unix(0, c.Date*1000)
	return fmt.Sprintf("CommitmentsOfTraders{Date: %s, Price: %f, OverallInterest: %d, "+
		"NonCommercialLong: %d, NonCommercialShort: %d, Unit: %s}", t.Format(time.RFC3339), c.Price,
		c.OverallInterest, c.NonCommercialLong, c.NonCommercialShort,
		c.Unit)
}

// CommitmentsOfTraders returns up to 4 years of commitments of traders.
//
// The commitments of traders report is released by the CFTC and provides a breakdown of each
// Tuesday's open interest.
func (c *Client) CommitmentsOfTraders(instrument string) ([]CommitmentsOfTraders, error) {
	instrument = strings.ToUpper(instrument)
	u, err := url.Parse("/labs/v1/commitments_of_traders")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("instrument", instrument)
	u.RawQuery = q.Encode()

	m := make(map[string][]CommitmentsOfTraders)
	if err = requestAndDecode(c, "GET", u.String(), nil, &m); err != nil {
		return nil, err
	}

	cot, ok := m[instrument]
	if !ok {
		return nil, fmt.Errorf("No CommitmentsOfTraders found for instrument %s", instrument)
	}
	return cot, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// OrderBooks

// Pricepoint defines the number of orders and positions at a certain price.
type PricePoint struct {
	Price          float64
	OrdersShort    float64 `json:"os"`
	OrdersLong     float64 `json:"ol"`
	PositionsShort float64 `json:"ps"`
	PositionsLong  float64 `json:"pl"`
}

func (pp PricePoint) String() string {
	return fmt.Sprintf("PricePoint{Price: %f, OrdersShort: %f, OrdersLong: %f, "+
		"PositionsShort: %f, PositionsLong: %f}", pp.Price, pp.OrdersShort,
		pp.OrdersLong, pp.PositionsShort, pp.PositionsLong)
}

// OrderBook represents the order book at a specific time.
type OrderBook struct {
	Timestamp   Time
	MarketPrice float64
	PricePoints []PricePoint
}

func (ob OrderBook) String() string {
	return fmt.Sprintf("OrderBook{Timestamp: %v, MarketPrice: %f, PricePoints %v}",
		ob.Timestamp, ob.MarketPrice, ob.PricePoints)
}

func (ob *OrderBook) UnmarshalJSON(data []byte) error {
	v := struct {
		MarketPrice *float64              `json:"rate"`
		PricePoints map[string]PricePoint `json:"price_points"`
	}{
		MarketPrice: &ob.MarketPrice,
		PricePoints: make(map[string]PricePoint),
	}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	ob.PricePoints = make([]PricePoint, 0, len(v.PricePoints))
	for priceStr, pp := range v.PricePoints {
		if price, err := strconv.ParseFloat(priceStr, 64); err != nil {
			return err
		} else {
			pp.Price = price
		}
		ob.PricePoints = append(ob.PricePoints, pp)
	}
	return nil
}

type OrderBooks []OrderBook

func (obs *OrderBooks) UnmarshalJSON(data []byte) error {
	m := make(map[string]OrderBook)
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	for timeStr, ob := range m {
		ob.Timestamp = Time(timeStr)
		*obs = append(*obs, ob)
	}
	return nil
}

// Orderbook returns historic order book data.
//
// See http://developer.oanda.com/docs/v1/forex-labs/#orderbook for further information.
func (c *Client) OrderBooks(instrument string, period Period) (OrderBooks, error) {
	instrument = strings.ToUpper(instrument)

	u, err := url.Parse("/labs/v1/orderbook_data")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("instrument", instrument)
	q.Set("period", strconv.Itoa(int(period)))
	u.RawQuery = q.Encode()

	obs := make(OrderBooks, 0)
	if err = getAndDecode(c, u.String(), &obs); err != nil {
		return nil, err
	}
	obs.Sort()
	return obs, nil
}

type pricePointSorter struct {
	pricePoints []PricePoint
}

type orderBookSorter struct {
	orderBooks OrderBooks
}

func (obs *orderBookSorter) Len() int { return len(obs.orderBooks) }

func (obs *orderBookSorter) Swap(i, j int) {
	obs.orderBooks[i], obs.orderBooks[j] = obs.orderBooks[j], obs.orderBooks[i]
}

func (obs *orderBookSorter) Less(i, j int) bool {
	return obs.orderBooks[i].Timestamp.UnixMicro() < obs.orderBooks[j].Timestamp.UnixMicro()
}

func (obs *OrderBooks) Sort() {
	sort.Sort(&orderBookSorter{*obs})
	for i := range *obs {
		(*obs)[i].Sort()
	}
}

func (pps *pricePointSorter) Len() int { return len(pps.pricePoints) }

func (pps *pricePointSorter) Swap(i, j int) {
	pps.pricePoints[i], pps.pricePoints[j] = pps.pricePoints[j], pps.pricePoints[i]
}

func (pps *pricePointSorter) Less(i, j int) bool {
	return pps.pricePoints[i].Price < pps.pricePoints[j].Price
}

func (ob *OrderBook) Sort() {
	pps := pricePointSorter{ob.PricePoints}
	sort.Sort(&pps)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// AutochartistPattern

type AutochartistArg interface {
	applyAutochartistArg(url.Values)
}

func (i Instrument) applyAutochartistArg(v url.Values) {
	v.Set("instrument", strings.ToUpper(string(i)))
}

func (p Period) applyAutochartistArg(v url.Values) {
	v.Set("period", strconv.Itoa(int(p)))
}

type Quality int

func (q Quality) applyAutochartistArg(v url.Values) {
	v.Set("quality", strconv.Itoa(int(q)))
}

type Direction string

func (d Direction) applyAutochartistArg(v url.Values) {
	v.Set("direction", string(d))
}

const (
	Bullish Direction = "bullish"
	Bearish Direction = "bearish"
)

type Stats struct {
	Total   int     `json:"total"`
	Percent float64 `json:"percent"`
	Correct int     `json:"correct"`
}

func (s Stats) String() string {
	return fmt.Sprintf("Stats{Total: %d, Percent: %v, Correct: %d}", s.Total, s.Percent,
		s.Correct)
}

type HistoricalStats struct {
	HourOfDay Stats `json:"hourofday"`
	Pattern   Stats `json:"pattern"`
	Symbol    Stats `json:"symbol"`
}

func (s HistoricalStats) String() string {
	return fmt.Sprintf("HistoricalStats{HourOfDay: %v, Pattern: %v, Symbol: %v}", s.HourOfDay,
		s.Pattern, s.Symbol)
}

type AutochartistSignalMeta struct {
	Completed int `json:"completed"`
	Scores    struct {
		Uniformity   int `json:"uniformity"`
		Quality      int `json:"quality"`
		Breakout     int `json:"breakout"`
		InitialTrend int `json:"initialtrend"`
		Clarity      int `json:"clarity"`
	}
	Probability     float64         `json:"probability"`
	Interval        int             `json:"interval"`
	Direction       int             `json:"direction"`
	Pattern         string          `json:"pattern"`
	Length          int             `json:"length"`
	HistoricalStats HistoricalStats `json:"historicalstats"`
	TrendType       string          `json:"trendtype"`
}

func (m AutochartistSignalMeta) String() string {
	return fmt.Sprintf("Meta{Completed: %v, Scores{Uniformity: %v, Quality: %v, Breakout: %v, "+
		"InitialTrend: %v, Clarity: %v}, Probability: %v, Interval: %v, Direction: %v, "+
		"Pattern: %v, Length: %v, HistoricalStats: %v, TrendType: %v}", m.Completed,
		m.Scores.Uniformity, m.Scores.Quality, m.Scores.Breakout, m.Scores.InitialTrend,
		m.Scores.Clarity, m.Probability, m.Interval, m.Direction, m.Pattern, m.Length,
		m.HistoricalStats, m.TrendType)
}

type Point struct {
	X0 int64   `json:"x0"`
	X1 int64   `json:"x1"`
	Y0 float64 `json:"y0"`
	Y1 float64 `json:"y1"`
}

func (p Point) String() string {
	x0, x1 := time.Unix(0, p.X0*1000), time.Unix(0, p.X1*1000)
	return fmt.Sprintf("Point{X0: %s, X1: %s, Y0: %v, Y1: %v}", x0.Format(time.RFC3339),
		x1.Format(time.RFC3339), p.Y0, p.Y1)
}

type Prediction struct {
	TimeTo    int64   `json:"timeto"`
	TimeFrom  int64   `json:"timefrom"`
	PriceHigh float64 `json:"pricehigh"`
	PriceLow  float64 `json:"pricelow"`
}

func (p Prediction) String() string {
	tt, tf := time.Unix(0, p.TimeTo*1000), time.Unix(0, p.TimeFrom*1000)
	return fmt.Sprintf("Prediction{TimeTo: %s, TimeFrom: %s, PriceHigh: %v, PriceLow: %v}",
		tt.Format(time.RFC3339), tf.Format(time.RFC3339), p.PriceHigh, p.PriceLow)
}

type AutochartistSignalData struct {
	PatternEndTime int64
	Points         struct {
		Resistance Point `json:"resistance"`
		Support    Point `json:"support"`
	} `json:"points"`
	Prediction Prediction `json:"prediction"`
}

func (d AutochartistSignalData) String() string {
	pet := time.Unix(0, d.PatternEndTime*1000)
	return fmt.Sprintf("Data{PatternEndTime: %s, Points{Resistance: %v, Support: %v}, "+
		"Prediction: %v}", pet.Format(time.RFC3339), d.Points.Resistance, d.Points.Support,
		d.Prediction)
}

type AutochartistSignal struct {
	Meta       AutochartistSignalMeta `json:"meta"`
	Id         Id                     `json:"id"`
	Instrument string                 `json:"instrument"`
	Type       string                 `json:"type"`
	Data       AutochartistSignalData `json:"data"`
}

func (s AutochartistSignal) String() string {
	return fmt.Sprintf("Signal{Id: %v, Instrument %v, Type: %v, Data: %v, Meta: %v}", s.Id,
		s.Instrument, s.Type, s.Data, s.Meta)
}

type AutochartistPattern struct {
	Signals  []AutochartistSignal `json:"signals"`
	Provider string               `json:"provider"`
}

func (p AutochartistPattern) String() string {
	return fmt.Sprintf("AutochartistPattern{Provider: %v, Signals: %v}", p.Provider,
		p.Signals)
}

// AutochartistPattern
func (c *Client) AutochartistPattern(arg ...AutochartistArg) (*AutochartistPattern, error) {
	u, err := url.Parse("/labs/v1/signal/autochartist")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("type", "chartpattern")
	for _, a := range arg {
		a.applyAutochartistArg(q)
	}
	u.RawQuery = q.Encode()

	pattern := AutochartistPattern{}
	if err := getAndDecode(c, u.String(), &pattern); err != nil {
		return nil, err
	}
	return &pattern, nil
}
