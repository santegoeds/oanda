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
)

type NewTradeArg interface {
	applyNewTradeArg(url.Values)
}

func (lb LowerBound) applyNewTradeArg(v url.Values) {
	optionalArgs(v).SetFloat("lowerBound", float64(lb))
}

func (ub UpperBound) applyNewTradeArg(v url.Values) {
	optionalArgs(v).SetFloat("upperBound", float64(ub))
}

func (sl StopLoss) applyNewTradeArg(v url.Values) {
	optionalArgs(v).SetFloat("stopLoss", float64(sl))
}

func (tp TakeProfit) applyNewTradeArg(v url.Values) {
	optionalArgs(v).SetFloat("takeProfit", float64(tp))
}

func (ts TrailingStop) applyNewTradeArg(v url.Values) {
	optionalArgs(v).SetFloat("trailingStop", float64(ts))
}

type TradesArg interface {
	applyTradesArg(url.Values)
}

func (c Count) applyTradesArg(v url.Values) {
	optionalArgs(v).SetInt("count", int(c))
}

func (mi MaxId) applyTradesArg(v url.Values) {
	optionalArgs(v).SetId("maxId", Id(mi))
}

func (i Instrument) applyTradesArg(v url.Values) {
	v.Set("instrument", string(i))
}

func (ids Ids) applyTradesArg(v url.Values) {
	optionalArgs(v).SetIdArray("ids", ids)
}

type ModifyTradeArg interface {
	applyModifyTradeArg(url.Values)
}

func (sl StopLoss) applyModifyTradeArg(v url.Values) {
	optionalArgs(v).SetFloat("stopLoss", float64(sl))
}

func (tp TakeProfit) applyModifyTradeArg(v url.Values) {
	optionalArgs(v).SetFloat("takeProfit", float64(tp))
}

func (ts TrailingStop) applyModifyTradeArg(v url.Values) {
	optionalArgs(v).SetFloat("trailingStop", float64(ts))
}

// Trade represents an open Oanda trade.
type Trade struct {
	TradeId        Id      `json:"id"`
	Units          int     `json:"units"`
	Instrument     string  `json:"instrument"`
	Side           string  `json:"side"`
	Price          float64 `json:"price"`
	Time           Time    `json:"time"`
	StopLoss       float64 `json:"stopLoss"`
	TakeProfit     float64 `json:"takeProfit"`
	TrailingStop   float64 `json:"trailingStop"`
	TrailingAmount float64 `json:"trailingAmount"`
}

// String implements the Stringer interface.
func (t *Trade) String() string {
	return fmt.Sprintf("Trade{TradeId: %d, Side: %s, Units: %d, Instrument: %s, Price: %f}",
		t.TradeId, t.Side, t.Units, t.Instrument, t.Price)
}

type Trades []Trade

// NewTrade submits a MarketOrder request to the Oanda servers. Supported OptionalArgs are
// UpperBound(), LowerBound(), StopLoss(), TakeProfit() and TrailingStop().
func (c *Client) NewTrade(side TradeSide, units int, instrument string,
	args ...NewTradeArg) (*Trade, error) {

	instrument = strings.ToUpper(instrument)

	data := url.Values{
		"type":       {"market"},
		"side":       {string(side)},
		"units":      {strconv.Itoa(units)},
		"instrument": {instrument},
	}

	for _, arg := range args {
		arg.applyNewTradeArg(data)
	}

	// FIXME: Replace this with a TradeCreatedResponse that mimics the structure that is actually
	// returned.
	t := &Trade{
		Side:       string(side),
		Units:      units,
		Instrument: instrument,
	}

	rspData := struct {
		Instrument   string  `json:"instrument"`
		Time         Time    `json:"time"`
		Price        float64 `json:"price"`
		TradeOpened  *Trade  `json:"tradeOpened"`
		TradeReduced *Trade  `json:"tradeReduced"`
	}{
		TradeOpened:  t,
		TradeReduced: t,
	}

	urlStr := fmt.Sprintf("/v1/accounts/%d/orders", c.accountId)
	if err := requestAndDecode(c, "POST", urlStr, data, &rspData); err != nil {
		return nil, err
	}

	t.Instrument = rspData.Instrument
	t.Time = rspData.Time
	t.Price = rspData.Price

	return t, nil
}

// Trade returns an open trade.
func (c *Client) Trade(tradeId Id) (*Trade, error) {
	t := Trade{}
	urlStr := fmt.Sprintf("/v1/accounts/%d/trades/%d", c.accountId, tradeId)
	if err := getAndDecode(c, urlStr, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// Trades returns a list of open trades that match the optional arguments.  Supported
// optional arguments are MaxId(), Count(), Instrument() and Ids().
func (c *Client) Trades(args ...TradesArg) (Trades, error) {
	urlStr := fmt.Sprintf("/v1/accounts/%d/trades", c.accountId)

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	for _, arg := range args {
		arg.applyTradesArg(q)
	}
	u.RawQuery = q.Encode()
	urlStr = u.String()

	rspData := struct {
		Trades Trades `json:"trades"`
	}{}
	if err = getAndDecode(c, urlStr, &rspData); err != nil {
		return nil, err
	}
	return rspData.Trades, nil
}

// ModifyTrade modifies an open trade.  Supported optional arguments are StopLoss(),
// TakeProfit(), TrailingStop()
func (c *Client) ModifyTrade(tradeId Id, arg ModifyTradeArg, args ...ModifyTradeArg) (*Trade, error) {
	data := url.Values{}
	arg.applyModifyTradeArg(data)
	for _, arg := range args {
		arg.applyModifyTradeArg(data)
	}
	t := Trade{}
	urlStr := fmt.Sprintf("/v1/accounts/%d/trades/%d", c.accountId, tradeId)
	if err := requestAndDecode(c, "PATCH", urlStr, data, &t); err != nil {
		return nil, err
	}

	return &t, nil
}

type CloseTradeResponse struct {
	TransactionId Id      `json:"id"`
	Price         float64 `json:"price"`
	Instrument    string  `json:"instrument"`
	Profit        float64 `json:"profit"`
	Side          string  `json:"side"`
	Time          Time    `json:"time"`
}

// CloseTrade closes an open trade.
func (c *Client) CloseTrade(tradeId Id) (*CloseTradeResponse, error) {
	ctr := CloseTradeResponse{}
	urlStr := fmt.Sprintf("/v1/accounts/%d/trades/%d", c.accountId, tradeId)
	if err := requestAndDecode(c, "DELETE", urlStr, nil, &ctr); err != nil {
		return nil, err
	}
	return &ctr, nil
}
