package oanda

import (
	"encoding/json"
	"fmt"
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

type CalendarEvent struct {
	data struct {
		Title     string
		Timestamp time.Time
		Unit      string
		Currency  string
		Forecast  float64
		Previous  float64
		Actual    float64
		Market    float64
	}
}

func (ce *CalendarEvent) Title() string        { return ce.data.Title }
func (ce *CalendarEvent) Timestamp() time.Time { return ce.data.Timestamp }
func (ce *CalendarEvent) Unit() string         { return ce.data.Unit }
func (ce *CalendarEvent) Currency() string     { return ce.data.Currency }
func (ce *CalendarEvent) Forecast() float64    { return ce.data.Forecast }
func (ce *CalendarEvent) Previous() float64    { return ce.data.Previous }
func (ce *CalendarEvent) Actual() float64      { return ce.data.Actual }
func (ce *CalendarEvent) Market() float64      { return ce.data.Market }

func (ce *CalendarEvent) UnmarshalJSON(data []byte) error {
	v := struct {
		Title     *string  `json:"title"`
		Timestamp int64    `json:"timestamp"`
		Unit      *string  `json:"unit"`
		Currency  *string  `json:"currency"`
		Forecast  *float64 `json:"forecast,string"`
		Previous  *float64 `json:"previous,string"`
		Actual    *float64 `json:"actual,string"`
		Market    *float64 `json:"market,string"`
	}{
		Title:    &ce.data.Title,
		Unit:     &ce.data.Unit,
		Currency: &ce.data.Unit,
		Forecast: &ce.data.Forecast,
		Previous: &ce.data.Previous,
		Actual:   &ce.data.Actual,
		Market:   &ce.data.Market,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	ce.data.Timestamp = time.Unix(v.Timestamp, 0)
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
	Instrument  string
	DisplayName string
	Ratios      []PositionRatio
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
			pr.Ratios[i].data.Timestamp = time.Unix(int64(ratio[0]), 0)
			pr.Ratios[i].data.LongRatio = ratio[1]
			pr.Ratios[i].data.ExchangeRate = ratio[2]
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

type Spread struct {
	data struct {
		Timestamp time.Time
		Spread    float64
	}
}

func (s *Spread) Timestamp() time.Time { return s.data.Timestamp }
func (s *Spread) Spread() float64      { return s.data.Spread }

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

type CommitmentsOfTraders struct {
	data struct {
		OverallInterest    int
		NonCommercialLong  int
		Price              float64
		Date               time.Time
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
