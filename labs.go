package oanda

import (
	"encoding/json"
	"strconv"
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
