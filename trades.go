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

type NewTradeArg interface {
	ApplyNewTradeArg(url.Values)
}

func (lb LowerBound) ApplyNewTradeArg(v url.Values) {
	optionalArgs(v).SetFloat("lowerBound", float64(lb))
}

func (ub UpperBound) ApplyNewTradeArg(v url.Values) {
	optionalArgs(v).SetFloat("upperBound", float64(ub))
}

func (sl StopLoss) ApplyNewTradeArg(v url.Values) {
	optionalArgs(v).SetFloat("stopLoss", float64(sl))
}

func (tp TakeProfit) ApplyNewTradeArg(v url.Values) {
	optionalArgs(v).SetFloat("takeProfit", float64(tp))
}

func (ts TrailingStop) ApplyNewTradeArg(v url.Values) {
	optionalArgs(v).SetFloat("trailingStop", float64(ts))
}

type TradesArg interface {
	ApplyTradesArg(url.Values)
}

func (c Count) ApplyTradesArg(v url.Values) {
	optionalArgs(v).SetInt("count", int(c))
}

func (mi MaxId) ApplyTradesArg(v url.Values) {
	optionalArgs(v).SetInt("maxId", int(mi))
}

func (i Instrument) ApplyTradesArg(v url.Values) {
	v.Set("instrument", string(i))
}

func (ids Ids) ApplyTradesArg(v url.Values) {
	optionalArgs(v).SetIntArray("ids", []int(ids))
}

type ModifyTradeArg interface {
	ApplyModifyTradeArg(url.Values)
}

func (sl StopLoss) ApplyModifyTradeArg(v url.Values) {
	optionalArgs(v).SetFloat("stopLoss", float64(sl))
}

func (tp TakeProfit) ApplyModifyTradeArg(v url.Values) {
	optionalArgs(v).SetFloat("takeProfit", float64(tp))
}

func (ts TrailingStop) ApplyModifyTradeArg(v url.Values) {
	optionalArgs(v).SetFloat("trailingStop", float64(ts))
}

// Trade represents an open Oanda trade.
type Trade struct {
	TradeId        int       `json:"id"`
	Units          int       `json:"units"`
	Instrument     string    `json:"instrument"`
	Side           string    `json:"side"`
	Price          float64   `json:"price"`
	Time           time.Time `json:"time"`
	StopLoss       float64   `json:"stopLoss"`
	TakeProfit     float64   `json:"takeProfit"`
	TrailingStop   float64   `json:"trailingStop"`
	TrailingAmount float64   `json:"trailingAmount"`
}

// String implements the Stringer interface.
func (t Trade) String() string {
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
		arg.ApplyNewTradeArg(data)
	}

	urlPath := fmt.Sprintf("/v1/accounts/%d/orders", c.AccountId)
	ctx, err := c.newContext("POST", c.getUrl(urlPath, "api"), data)
	if err != nil {
		return nil, err
	}

	t := &Trade{
		Side:       string(side),
		Units:      units,
		Instrument: instrument,
	}

	rspData := struct {
		Instrument   string    `json:"instrument"`
		Time         time.Time `json:"time"`
		Price        float64   `json:"price"`
		TradeOpened  *Trade    `json:"tradeOpened"`
		TradeReduced *Trade    `json:"tradeReduced"`
	}{
		TradeOpened:  t,
		TradeReduced: t,
	}
	if _, err = ctx.Decode(&rspData); err != nil {
		return nil, err
	}

	t.Instrument = rspData.Instrument
	t.Time = rspData.Time
	t.Price = rspData.Price

	return t, nil
}

// Trade returns an open trade.
func (c *Client) Trade(tradeId int) (*Trade, error) {
	u := c.getUrl(fmt.Sprintf("/v1/accounts/%d/trades/%d", c.AccountId, tradeId), "api")
	ctx, err := c.newContext("GET", u, nil)
	if err != nil {
		return nil, err
	}

	t := Trade{}
	if _, err = ctx.Decode(&t); err != nil {
		return nil, err
	}

	return &t, nil
}

// Trades returns a list of open trades that match the optional arguments.  Supported
// optional arguments are MaxId(), Count(), Instrument() and Ids().
func (c *Client) Trades(args ...TradesArg) (Trades, error) {
	u := c.getUrl(fmt.Sprintf("/v1/accounts/%d/trades", c.AccountId), "api")
	q := u.Query()
	for _, arg := range args {
		arg.ApplyTradesArg(q)
	}
	u.RawQuery = q.Encode()

	ctx, err := c.newContext("GET", u, nil)
	if err != nil {
		return nil, err
	}

	rspData := struct {
		Trades Trades `json:"trades"`
	}{}

	if _, err = ctx.Decode(&rspData); err != nil {
		return nil, err
	}

	return rspData.Trades, nil
}

// ModifyTrade modifies an open trade.  Supported optional arguments are StopLoss(),
// TakeProfit(), TrailingStop()
func (c *Client) ModifyTrade(tradeId int,
	arg ModifyTradeArg, args ...ModifyTradeArg) (*Trade, error) {

	data := url.Values{}
	arg.ApplyModifyTradeArg(data)
	for _, arg := range args {
		arg.ApplyModifyTradeArg(data)
	}

	u := c.getUrl(fmt.Sprintf("/v1/accounts/%d/trades/%d", c.AccountId, tradeId), "api")
	ctx, err := c.newContext("PATCH", u, data)
	if err != nil {
		return nil, err
	}

	t := Trade{}
	if _, err = ctx.Decode(&t); err != nil {
		return nil, err
	}

	return &t, nil
}

type CloseTradeResponse struct {
	TransactionId int       `json:"id"`
	Price         float64   `json:"price"`
	Instrument    string    `json:"instrument"`
	Profit        float64   `json:"profit"`
	Side          string    `json:"side"`
	Time          time.Time `json:"time"`
}

// CloseTrade closes an open trade.
func (c *Client) CloseTrade(tradeId int) (*CloseTradeResponse, error) {
	u := c.getUrl(fmt.Sprintf("/v1/accounts/%d/trades/%d", c.AccountId, tradeId), "api")
	ctx, err := c.newContext("DELETE", u, nil)
	if err != nil {
		return nil, err
	}

	ctr := CloseTradeResponse{}
	if _, err = ctx.Decode(&ctr); err != nil {
		return nil, err
	}

	return &ctr, nil
}