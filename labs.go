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
	Title     string    `json:"title"`
	Timestamp time.Time `json:"timestamp"`
	Unit      string    `json:"unit"`
	Currency  string    `json:"currency"`
	Forecast  float64   `json:"forecast,string"`
	Previous  float64   `json:"previous,string"`
	Actual    float64   `json:"actual,string"`
	Market    float64   `json:"market,string"`
}

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
		Title:    &ce.Title,
		Unit:     &ce.Unit,
		Currency: &ce.Unit,
		Forecast: &ce.Forecast,
		Previous: &ce.Previous,
		Actual:   &ce.Actual,
		Market:   &ce.Market,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	ce.Timestamp = time.Unix(v.Timestamp, 0)
	return nil
}

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
	Timestamp    time.Time
	LongRatio    float64
	ExchangeRate float64
}

type PositionRatios struct {
	Instrument  string
	DisplayName string
	Ratios      []PositionRatio
}

func (c *Client) HistoricPositionRatios(instrument string, period Period) (*PositionRatios, error) {
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

	pr := PositionRatios{
		Instrument: instrument,
	}

	v := struct {
		Data map[string]struct {
			Data  [][]float64 `json:"data"`
			Label string      `json:"label"`
		} `json:"data"`
	}{}

	dec := json.NewDecoder(rsp.Body)
	if err = dec.Decode(&v); err != nil {
		return nil, err
	}
	data, ok := v.Data[instrument]
	if !ok || len(data.Data) == 0 {
		return nil, fmt.Errorf("No HistoricPositionRatios found for instrument %s", instrument)
	}

	pr.DisplayName = data.Label
	pr.Ratios = make([]PositionRatio, len(data.Data))
	for i, ratio := range data.Data {
		pr.Ratios[i].Timestamp = time.Unix(int64(ratio[0]), 0)
		pr.Ratios[i].LongRatio = ratio[1]
		pr.Ratios[i].ExchangeRate = ratio[2]
	}

	return &pr, nil
}
