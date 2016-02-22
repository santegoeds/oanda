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
	"strings"
)

type (
	Ids []Id
)

type Position struct {
	Side       string  `json:"side"`
	Instrument string  `json:"instrument"`
	Units      int     `json:"units"`
	AvgPrice   float64 `json:"avgPrice"`
}

// String implements the fmt.Stringer interface.
func (p Position) String() string {
	return fmt.Sprintf("Position{Side: %s, Instrument: %s, Units: %d, AvgPrice: %f}", p.Side,
		p.Instrument, p.Units, p.AvgPrice)
}

type PositionCloseResponse struct {
	// Ids are the transaction ids that are created as a result of closing the position.
	TranIds    Ids    `json:"ids"`
	Instrument string `json:"instrument"`
	TotalUnits int    `json:"totalUnits"`
}

type Positions []Position

// Positions returns all positions for the selected account.
func (c *Client) Positions() (Positions, error) {
	urlStr := fmt.Sprintf("/v1/accounts/%d/positions", c.accountId)
	positions := struct {
		Positions Positions `json:"positions"`
	}{}
	if err := getAndDecode(c, urlStr, &positions); err != nil {
		return nil, err
	}
	return positions.Positions, nil
}

// Position returns the position for the selected account and instrument.
func (c *Client) Position(instrument string) (*Position, error) {
	instrument = strings.ToUpper(instrument)
	urlStr := fmt.Sprintf("/v1/accounts/%d/positions/%s", c.accountId, instrument)
	p := Position{}
	if err := getAndDecode(c, urlStr, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ClosePosition closes an existing position.
func (c *Client) ClosePosition(instrument string) (*PositionCloseResponse, error) {
	instrument = strings.ToUpper(instrument)
	pcr := PositionCloseResponse{}
	urlStr := fmt.Sprintf("/v1/accounts/%d/positions/%s", c.accountId, instrument)
	if err := requestAndDecode(c, "DELETE", urlStr, nil, &pcr); err != nil {
		return nil, err
	}
	return &pcr, nil
}
