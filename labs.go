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

type Period int

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
	data struct {
		Title     string    `json:"title"`
		Timestamp posixTime `json:"timestamp"`
		Unit      string    `json:"unit"`
		Currency  string    `json:"currency"`
		Forecast  float64   `json:"forecast,string"`
		Previous  float64   `json:"previous,string"`
		Actual    float64   `json:"actual,string"`
		Market    float64   `json:"market,string"`
	}
}

func (ce *CalendarEvent) Title() string        { return ce.data.Title }
func (ce *CalendarEvent) Timestamp() time.Time { return ce.data.Timestamp.Time }
func (ce *CalendarEvent) Unit() string         { return ce.data.Unit }
func (ce *CalendarEvent) Currency() string     { return ce.data.Currency }
func (ce *CalendarEvent) Forecast() float64    { return ce.data.Forecast }
func (ce *CalendarEvent) Previous() float64    { return ce.data.Previous }
func (ce *CalendarEvent) Actual() float64      { return ce.data.Actual }
func (ce *CalendarEvent) Market() float64      { return ce.data.Market }

func (ce CalendarEvent) String() string {
	return fmt.Sprintf("CalendarEvent{Title: %s, Timestamp: %s, Unit: %s, Currency: %s, "+
		"Forecast: %v, Previous: %v, Actual: %v, Market: %v}", ce.data.Title,
		ce.data.Timestamp.Format(time.RFC3339), ce.data.Unit, ce.data.Currency, ce.data.Forecast,
		ce.data.Previous, ce.data.Actual, ce.data.Market)
}

func (ce *CalendarEvent) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &ce.data); err != nil {
		return err
	}
	return nil
}

// Calendar returns and array of economic calendar events associated with an instrument. Events
// can include economic indicator data or they can solely be be news about important meetings.
//
// See http://developer.oanda.com/docs/v1/forex-labs/#calendar for further information.
func (c *Client) Calendar(instrument string, period Period) ([]CalendarEvent, error) {
	instrument = strings.ToUpper(instrument)
	req, err := c.NewRequest("GET", "/labs/v1/calendar", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("instrument", instrument)
	q.Set("period", strconv.Itoa(int(period)))
	req.URL.RawQuery = q.Encode()

	rsp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	dec := json.NewDecoder(rsp.Body)
	ces := []CalendarEvent{}
	if err := dec.Decode(&ces); err != nil {
		return nil, err
	}
	return ces, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// PositionRatios

type PositionRatio struct {
	data struct {
		Timestamp    time.Time
		LongRatio    float64
		ExchangeRate float64
	}
}

func (pr *PositionRatio) Timestamp() time.Time  { return pr.data.Timestamp }
func (pr *PositionRatio) LongRatio() float64    { return pr.data.LongRatio }
func (pr *PositionRatio) ShortRatio() float64   { return 100.0 - pr.LongRatio() }
func (pr *PositionRatio) ExchangeRate() float64 { return pr.data.ExchangeRate }

type PositionRatios struct {
	data struct {
		Instrument  string
		DisplayName string
		Ratios      []PositionRatio
	}
}

func (pr *PositionRatios) Instrument() string      { return pr.data.Instrument }
func (pr *PositionRatios) DisplayName() string     { return pr.data.DisplayName }
func (pr *PositionRatios) Ratios() []PositionRatio { return pr.data.Ratios }

func (pr PositionRatios) String() string {
	return fmt.Sprintf("PositionRatios{Instrument: %s, DisplayName: %s, Ratios: %v}",
		pr.data.Instrument, pr.data.DisplayName, pr.data.Ratios)
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
		pr.data.Instrument = instrument
		pr.data.DisplayName = data.Label
		pr.data.Ratios = make([]PositionRatio, len(data.Data))
		for i, ratio := range data.Data {
			pr.data.Ratios[i].data.Timestamp = time.Unix(int64(ratio[0]), 0)
			pr.data.Ratios[i].data.LongRatio = ratio[1]
			pr.data.Ratios[i].data.ExchangeRate = ratio[2]
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
	req, err := c.NewRequest("GET", "/labs/v1/historical_position_ratios", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("instrument", instrument)
	q.Set("period", strconv.Itoa(int(period)))
	req.URL.RawQuery = q.Encode()

	rsp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	pr := PositionRatios{}
	dec := json.NewDecoder(rsp.Body)
	if err = dec.Decode(&pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Spreads

type Spread struct {
	data struct {
		Timestamp time.Time
		Spread    float64
	}
}

func (s *Spread) Timestamp() time.Time { return s.data.Timestamp }
func (s *Spread) Spread() float64      { return s.data.Spread }

func (s Spread) String() string {
	return fmt.Sprintf("Spread{Timestamp: %s, Spread: %f}", s.data.Timestamp.Format(time.RFC3339),
		s.data.Spread)
}

func (s *Spread) UnmarshalJSON(data []byte) error {
	v := []float64{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	s.data.Timestamp = time.Unix(int64(v[0]), 0)
	s.data.Spread = v[1]
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
	req, err := c.NewRequest("GET", "/labs/v1/spreads", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("instrument", instrument)
	q.Set("period", strconv.Itoa(int(period)))
	if unique {
		q.Set("unique", "1")
	} else {
		q.Set("unique", "0")
	}
	req.URL.RawQuery = q.Encode()

	rsp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	s := Spreads{}
	dec := json.NewDecoder(rsp.Body)
	if err = dec.Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// CommitmentsOfTraders

type CommitmentsOfTraders struct {
	data struct {
		Date               time.Time
		Price              float64
		OverallInterest    int
		NonCommercialLong  int
		NonCommercialShort int
		Unit               string
	}
}

func (c *CommitmentsOfTraders) OverallInterest() int    { return c.data.OverallInterest }
func (c *CommitmentsOfTraders) NonCommercialLong() int  { return c.data.NonCommercialLong }
func (c *CommitmentsOfTraders) Price() float64          { return c.data.Price }
func (c *CommitmentsOfTraders) Date() time.Time         { return c.data.Date }
func (c *CommitmentsOfTraders) NonCommercialShort() int { return c.data.NonCommercialShort }
func (c *CommitmentsOfTraders) Unit() string            { return c.data.Unit }

func (c CommitmentsOfTraders) String() string {
	return fmt.Sprintf("CommitmentsOfTraders{Date: %s, Price: %f, OverallInterest: %d, "+
		"NonCommercialLong: %d, NonCommercialShort: %d, Unit: %s}", c.data.Date.Format(time.RFC3339),
		c.data.Price, c.data.OverallInterest, c.data.NonCommercialLong, c.data.NonCommercialShort,
		c.data.Unit)
}

func (c *CommitmentsOfTraders) UnmarshalJSON(data []byte) error {
	v := struct {
		OverallInterest    *int     `json:"oi,string"`
		NonCommercialLong  *int     `json:"ncl,string"`
		Price              *float64 `json:"price,string"`
		Date               int
		NonCommercialShort *int    `json:"ncs,string"`
		Unit               *string `json:"unit"`
	}{
		OverallInterest:    &c.data.OverallInterest,
		NonCommercialLong:  &c.data.NonCommercialLong,
		Price:              &c.data.Price,
		NonCommercialShort: &c.data.NonCommercialShort,
		Unit:               &c.data.Unit,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	c.data.Date = time.Unix(int64(v.Date), 0)
	return nil
}

// CommitmentsOfTraders returns up to 4 years of commitments of traders.
//
// The commitments of traders report is released by the CFTC and provides a breakdown of each
// Tuesday's open interest.
func (c *Client) CommitmentsOfTraders(instrument string) ([]CommitmentsOfTraders, error) {
	instrument = strings.ToUpper(instrument)
	req, err := c.NewRequest("GET", "/labs/v1/commitments_of_traders", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("instrument", instrument)
	req.URL.RawQuery = q.Encode()

	rsp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	m := make(map[string][]CommitmentsOfTraders)
	dec := json.NewDecoder(rsp.Body)
	if err = dec.Decode(&m); err != nil {
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
	data struct {
		Price          float64
		OrdersShort    float64 `json:"os"`
		OrdersLong     float64 `json:"ol"`
		PositionsShort float64 `json:"ps"`
		PositionsLong  float64 `json:"pl"`
	}
}

func (pp *PricePoint) Price() float64          { return pp.data.Price }
func (pp *PricePoint) OrdersShort() float64    { return pp.data.OrdersShort }
func (pp *PricePoint) OrdersLong() float64     { return pp.data.OrdersLong }
func (pp *PricePoint) PositionsShort() float64 { return pp.data.PositionsShort }
func (pp *PricePoint) PositionsLong() float64  { return pp.data.PositionsLong }

func (pp PricePoint) String() string {
	return fmt.Sprintf("PricePoint{Price: %f, OrdersShort: %f, OrdersLong: %f, "+
		"PositionsShort: %f, PositionsLong: %f}", pp.data.Price, pp.data.OrdersShort,
		pp.data.OrdersLong, pp.data.PositionsShort, pp.data.PositionsLong)
}

func (pp *PricePoint) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &pp.data)
}

// OrderBook represents the order book at a specific time.
type OrderBook struct {
	data struct {
		Timestamp   time.Time
		MarketPrice float64
		PricePoints []PricePoint
	}
}

func (ob *OrderBook) Timestamp() time.Time      { return ob.data.Timestamp }
func (ob *OrderBook) MarketPrice() float64      { return ob.data.MarketPrice }
func (ob *OrderBook) PricePoints() []PricePoint { return ob.data.PricePoints }

func (ob OrderBook) String() string {
	return fmt.Sprintf("OrderBook{Timestamp: %s, MarketPrice: %f, PricePoints %v}",
		ob.data.Timestamp.Format(time.RFC3339), ob.data.MarketPrice, ob.data.PricePoints)
}

func (ob *OrderBook) UnmarshalJSON(data []byte) error {
	v := struct {
		MarketPrice *float64              `json:"rate"`
		PricePoints map[string]PricePoint `json:"price_points"`
	}{
		MarketPrice: &ob.data.MarketPrice,
		PricePoints: make(map[string]PricePoint),
	}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	ob.data.PricePoints = make([]PricePoint, 0, len(v.PricePoints))
	for priceStr, pp := range v.PricePoints {
		if price, err := strconv.ParseFloat(priceStr, 64); err != nil {
			return err
		} else {
			pp.data.Price = price
		}
		ob.data.PricePoints = append(ob.data.PricePoints, pp)
	}
	return nil
}

type OrderBooks []OrderBook

func (obs *OrderBooks) UnmarshalJSON(data []byte) error {
	if *obs == nil {
		*obs = make(OrderBooks, 0)
	}
	m := make(map[string]OrderBook)
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	for timeStr, ob := range m {
		if unixTime, err := strconv.Atoi(timeStr); err != nil {
			return err
		} else {
			ob.data.Timestamp = time.Unix(int64(unixTime), 0)
		}
		*obs = append(*obs, ob)
	}
	return nil
}

// Orderbook returns historic order book data.
//
// See http://developer.oanda.com/docs/v1/forex-labs/#orderbook for further information.
func (c *Client) OrderBooks(instrument string, period Period) (OrderBooks, error) {
	instrument = strings.ToUpper(instrument)
	req, err := c.NewRequest("GET", "/labs/v1/orderbook_data", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("instrument", instrument)
	q.Set("period", strconv.Itoa(int(period)))
	req.URL.RawQuery = q.Encode()

	rsp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	obs := make(OrderBooks, 0)
	dec := json.NewDecoder(rsp.Body)
	if err = dec.Decode(&obs); err != nil {
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
	return obs.orderBooks[i].data.Timestamp.After(obs.orderBooks[j].data.Timestamp)
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
	return pps.pricePoints[i].data.Price < pps.pricePoints[j].data.Price
}

func (ob *OrderBook) Sort() {
	pps := pricePointSorter{ob.data.PricePoints}
	sort.Sort(&pps)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// AutochartistPattern

type AutochartistPatternArg interface {
	applyAutochartistPatternArg(url.Values)
}

func (i Instrument) applyAutochartistPatternArg(v url.Values) {
	v.Set("instrument", strings.ToUpper(string(i)))
}

func (p Period) applyAutochartistPatternArg(v url.Values) {
	v.Set("period", strconv.Itoa(int(p)))
}

type Quality int

func (q Quality) applyAutochartistPatternArg(v url.Values) {
	v.Set("quality", strconv.Itoa(int(q)))
}

type Direction string

func (d Direction) applyAutochartistPatternArg(v url.Values) {
	v.Set("direction", string(d))
}

const (
	Bullish Direction = "bullish"
	Bearish Direction = "bearish"
)

type Stats struct {
	data struct {
		Total   int     `json:"total"`
		Percent float64 `json:"percent"`
		Correct int     `json:"correct"`
	}
}

func (s *Stats) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &s.data)
}

func (s *Stats) Total() int       { return s.data.Total }
func (s *Stats) Percent() float64 { return s.data.Percent }
func (s *Stats) Correct() int     { return s.data.Correct }

func (s Stats) String() string {
	return fmt.Sprintf("Stats{Total: %d, Percent: %v, Correct: %d}", s.data.Total, s.data.Percent,
		s.data.Correct)
}

type HistoricalStats struct {
	data struct {
		HourOfDay Stats `json:"hourofday"`
		Pattern   Stats `json:"pattern"`
		Symbol    Stats `json:"symbol"`
	}
}

func (s HistoricalStats) String() string {
	return fmt.Sprintf("HistoricalStats{HourOfDay: %v, Pattern: %v, Symbol: %v}", s.data.HourOfDay,
		s.data.Pattern, s.data.Symbol)
}

func (hs *HistoricalStats) HourOfDay() Stats { return hs.data.HourOfDay }
func (hs *HistoricalStats) Pattern() Stats   { return hs.data.Pattern }
func (hs *HistoricalStats) Symbol() Stats    { return hs.data.Symbol }

func (hs *HistoricalStats) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &hs.data)
}

type AutochartistSignalMeta struct {
	data struct {
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
}

func (m AutochartistSignalMeta) String() string {
	return fmt.Sprintf("Meta{Completed: %v, Scores{Uniformity: %v, Quality: %v, Breakout: %v, "+
		"InitialTrend: %v, Clarity: %v}, Probability: %v, Interval: %v, Direction: %v, "+
		"Pattern: %v, Length: %v, HistoricalStats: %v, TrendType: %v}", m.data.Completed,
		m.data.Scores.Uniformity, m.data.Scores.Quality, m.data.Scores.Breakout,
		m.data.Scores.InitialTrend, m.data.Scores.Clarity, m.data.Probability, m.data.Interval,
		m.data.Direction, m.data.Pattern, m.data.Length, m.data.HistoricalStats, m.data.TrendType)
}

func (m *AutochartistSignalMeta) Completed() int                    { return m.data.Completed }
func (m *AutochartistSignalMeta) UniformityScore() int              { return m.data.Scores.Uniformity }
func (m *AutochartistSignalMeta) QualityScore() int                 { return m.data.Scores.Quality }
func (m *AutochartistSignalMeta) BreakoutScore() int                { return m.data.Scores.Breakout }
func (m *AutochartistSignalMeta) InitialTrendScore() int            { return m.data.Scores.InitialTrend }
func (m *AutochartistSignalMeta) ClarityScore() int                 { return m.data.Scores.Clarity }
func (m *AutochartistSignalMeta) Probability() float64              { return m.data.Probability }
func (m *AutochartistSignalMeta) Interval() int                     { return m.data.Interval }
func (m *AutochartistSignalMeta) Direction() int                    { return m.data.Direction }
func (m *AutochartistSignalMeta) Pattern() string                   { return m.data.Pattern }
func (m *AutochartistSignalMeta) Length() int                       { return m.data.Length }
func (m *AutochartistSignalMeta) HistoricalStats() *HistoricalStats { return &m.data.HistoricalStats }
func (m *AutochartistSignalMeta) TrendType() string                 { return m.data.TrendType }

func (m *AutochartistSignalMeta) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &m.data)
}

type Point struct {
	data struct {
		X0 posixTime `json:"x0"`
		X1 posixTime `json:"x1"`
		Y0 float64   `json:"y0"`
		Y1 float64   `json:"y1"`
	}
}

func (p Point) String() string {
	return fmt.Sprintf("Point{X0: %s, X1: %s, Y0: %v, Y1: %v}", p.data.X0.Format(time.RFC3339),
		p.data.X1.Format(time.RFC3339), p.data.Y0, p.data.Y1)
}

func (p *Point) X0() time.Time { return p.data.X0.Time }
func (p *Point) X1() time.Time { return p.data.X1.Time }
func (p *Point) Y0() float64   { return p.data.Y0 }
func (p *Point) Y1() float64   { return p.data.Y1 }

func (p *Point) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &p.data)
}

type Prediction struct {
	data struct {
		TimeTo    posixTime `json:"timeto"`
		TimeFrom  posixTime `json:"timefrom"`
		PriceHigh float64   `json:"pricehigh"`
		PriceLow  float64   `json:"pricelow"`
	}
}

func (p Prediction) String() string {
	return fmt.Sprintf("Prediction{TimeTo: %s, TimeFrom: %s, PriceHigh: %v, PriceLow: %v}",
		p.data.TimeTo.Format(time.RFC3339), p.data.TimeFrom.Format(time.RFC3339),
		p.data.PriceHigh, p.data.PriceLow)
}

func (p Prediction) TimeTo() time.Time   { return p.data.TimeTo.Time }
func (p Prediction) TimeFrom() time.Time { return p.data.TimeFrom.Time }
func (p Prediction) PriceHigh() float64  { return p.data.PriceHigh }
func (p Prediction) PriceLow() float64   { return p.data.PriceLow }

func (p *Prediction) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &p.data)
}

type AutochartistSignalData struct {
	data struct {
		PatternEndTime posixTime
		Points         struct {
			Resistance Point `json:"resistance"`
			Support    Point `json:"support"`
		} `json:"points"`
		Prediction Prediction `json:"prediction"`
	}
}

func (d AutochartistSignalData) String() string {
	return fmt.Sprintf("Data{PatternEndTime: %s, Points{Resistance: %v, Support: %v}, "+
		"Prediction: %v}", d.data.PatternEndTime.Format(time.RFC3339), d.data.Points.Resistance,
		d.data.Points.Support, d.data.Prediction)
}

func (d *AutochartistSignalData) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &d.data)
}

func (d *AutochartistSignalData) PatternEndTime() time.Time { return d.data.PatternEndTime.Time }
func (d *AutochartistSignalData) Resistance() Point         { return d.data.Points.Resistance }
func (d *AutochartistSignalData) Support() Point            { return d.data.Points.Support }
func (d *AutochartistSignalData) Prediction() Prediction    { return d.data.Prediction }

type AutochartistSignal struct {
	data struct {
		Meta       AutochartistSignalMeta `json:"meta"`
		Id         int                    `json:"id"`
		Instrument string                 `json:"instrument"`
		Type       string                 `json:"type"`
		Data       AutochartistSignalData `json:"data"`
	}
}

func (s AutochartistSignal) String() string {
	return fmt.Sprintf("Signal{Id: %v, Instrument %v, Type: %v, Data: %v, Meta: %v}", s.data.Id,
		s.data.Instrument, s.data.Type, s.data.Data, s.data.Meta)
}

func (s *AutochartistSignal) Meta() *AutochartistSignalMeta { return &s.data.Meta }
func (s *AutochartistSignal) Id() int                       { return s.data.Id }
func (s *AutochartistSignal) Instrument() string            { return s.data.Instrument }
func (s *AutochartistSignal) Type() string                  { return s.data.Type }
func (s *AutochartistSignal) Data() *AutochartistSignalData { return &s.data.Data }

func (s *AutochartistSignal) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &s.data)
}

type autochartistPattern struct {
	Signals  []AutochartistSignal `json:"signals"`
	Provider string               `json:"provider"`
}

type AutochartistPattern struct {
	data autochartistPattern
}

func (p *AutochartistPattern) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &p.data)
}

func (p AutochartistPattern) String() string {
	return fmt.Sprintf("AutochartistPattern{Provider: %v, Signals: %v}", p.data.Provider,
		p.data.Signals)
}

func (p *AutochartistPattern) Signals() []AutochartistSignal { return p.data.Signals }
func (p *AutochartistPattern) Provider() string              { return p.data.Provider }

// AutochartistPattern
func (c *Client) AutochartistPattern(arg ...AutochartistPatternArg) (*AutochartistPattern, error) {
	u, err := url.Parse("/labs/v1/signal/autochartist")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("type", "chartpattern")
	for _, a := range arg {
		a.applyAutochartistPatternArg(q)
	}
	u.RawQuery = q.Encode()

	pattern := AutochartistPattern{}
	if err := getAndDecode(c, u.String(), &pattern); err != nil {
		return nil, err
	}
	return &pattern, nil
}
