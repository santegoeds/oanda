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

type (
	TradeSide string
	OrderType string
)

const (
	Buy  TradeSide = "buy"
	Sell TradeSide = "sell"

	MarketIfTouched OrderType = "marketIfTouched"
	Limit           OrderType = "limit"
	Stop            OrderType = "stop"
)

type Order struct {
	OrderId        int       `json:"id"`
	Units          int       `json:"units"`
	Instrument     string    `json:"instrument"`
	Side           string    `json:"side"`
	Price          float64   `json:"price"`
	Time           time.Time `json:"time"`
	StopLoss       float64   `json:"stopLoss"`
	TakeProfit     float64   `json:"takeProfit"`
	TrailingStop   float64   `json:"trailingStop"`
	TrailingAmount float64   `json:"trailingAmount"`
	OrderType      string    `json:"type"`
	Expiry         time.Time `json:"expiry"`
	UpperBound     float64   `json:"upperBound"`
	LowerBound     float64   `json:"lowerBound"`
}

// String implements the fmt.Stringer interface.
func (o *Order) String() string {
	return fmt.Sprintf("Order{OrderId: %d, Side: %s, Units: %d, Instrument: %s}", o.OrderId, o.Side,
		o.Units, o.Instrument)
}

// LowerBound is an optional argument for Client methods NewOrder(), ModifyOrder() and
// NewTrade().
type LowerBound float64

// UpperBound is an optional argument for Client methods NewOrder(), ModifyOrder() and
// NewTrade().
type UpperBound float64

// StopLoss is an optional argument for Client methods  NewOrder(), ModifyOrder(), NewTrade()
// and ModifyTrade().
type StopLoss float64

// TakeProfit is an optional argument for Client methods NewOrder(), ModifyOrder(), NewTrade(),
// and ModifyTrade().
type TakeProfit float64

// TrailingStop is an optional argument for Client methods NewOrder(), ModifyOrder(), NewTrade()
// and ModifyTrade().
type TrailingStop float64

// NewOrderArg represents an optional argument for method NewOrder. Types that implement the
// interface are LowerBound, UpperBound, StopLoss, TakeProfit and TrailingStop.
type NewOrderArg interface {
	applyNewOrderArg(url.Values)
}

func (lb LowerBound) applyNewOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("lowerBound", float64(lb))
}

func (ub UpperBound) applyNewOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("upperBound", float64(ub))
}

func (sl StopLoss) applyNewOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("stopLoss", float64(sl))
}

func (tp TakeProfit) applyNewOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("takeProfit", float64(tp))
}

func (ts TrailingStop) applyNewOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("trailingStop", float64(ts))
}

// NewOrder creates and submits a new order.
func (c *Client) NewOrder(orderType OrderType, side TradeSide, units int, instrument string,
	price float64, expiry time.Time, args ...NewOrderArg) (*Order, error) {

	instrument = strings.ToUpper(instrument)

	o := Order{
		Side:       string(side),
		Units:      units,
		Instrument: instrument,
		Price:      price,
		OrderType:  string(orderType),
		Expiry:     expiry,
	}
	data := url.Values{
		"type":       {string(orderType)},
		"side":       {string(side)},
		"units":      {strconv.Itoa(units)},
		"instrument": {instrument},
		"price":      {strconv.FormatFloat(price, 'f', -1, 64)},
		"expiry":     {expiry.UTC().Format(time.RFC3339)},
	}
	for _, arg := range args {
		arg.applyNewOrderArg(data)
	}

	rspData := struct {
		ApiError
		Instrument  string    `json:"instrument"`
		Time        time.Time `json:"time"`
		Price       float64   `json:"price"`
		OrderOpened *Order    `json:"orderOpened"`
	}{
		OrderOpened: &o,
	}
	urlStr := fmt.Sprintf("/v1/accounts/%d/orders", c.accountId)
	if err := requestAndDecode(c, "POST", urlStr, data, &rspData); err != nil {
		return nil, err
	}
	o.Instrument = rspData.Instrument
	o.Time = rspData.Time
	o.Price = rspData.Price
	return &o, nil
}

// Order returns information about an existing order.
func (c *Client) Order(orderId int) (*Order, error) {
	o := struct {
		ApiError
		Order
	}{}
	urlStr := fmt.Sprintf("/v1/accounts/%d/orders/%d", c.accountId, orderId)
	if err := getAndDecode(c, urlStr, &o); err != nil {
		return nil, err
	}
	return &o.Order, nil
}

// MaxId is an optional argument for Client methods Events(), Orders() and Trades().
type MaxId int

// Count is an optional argument for Client methods Events(), Orders(), /MidpriceCandles(),
// BidAskCandles() and Trades().
type Count int

// Instrument is an optional argument for Client methods Events(), Orders() and Trades().
type Instrument string

// OrderArgs represents an optional argument for method Orders. Types that implement the interface
// are MaxId, Count and Instrument.
type OrdersArg interface {
	applyOrdersArg(url.Values)
}

func (mi MaxId) applyOrdersArg(v url.Values) {
	optionalArgs(v).SetInt("maxId", int(mi))
}

func (cnt Count) applyOrdersArg(v url.Values) {
	optionalArgs(v).SetInt("count", int(cnt))
}

func (in Instrument) applyOrdersArg(v url.Values) {
	v.Set("instrument", string(in))
}

// Orders returns an array with all orders that match the optional arguments (if any). Supported
// OrdersArg are MaxId, Count and Instrument.
func (c *Client) Orders(args ...OrdersArg) ([]Order, error) {
	u, err := url.Parse(fmt.Sprintf("/v1/accounts/%d/orders", c.accountId))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	for _, arg := range args {
		arg.applyOrdersArg(q)
	}
	u.RawQuery = q.Encode()

	rsp := struct {
		ApiError
		Orders []Order `json:"orders"`
	}{}
	if err := getAndDecode(c, u.String(), &rsp); err != nil {
		return nil, err
	}
	return rsp.Orders, nil
}

// Units is an optional argument for Client method ModifyOrder().
type Units int

// Expiry is an optional argument for Client method ModifyOrder().
type Expiry time.Time

// Price is an optional argument for Client method ModifyOrder().
type Price float64

// ModifyOrderArg represents an opional argument for method ModifyOrder. Types that implement
// the interface are Units, Price, Expiry, LowerBound, UpperBound, StopLoss, TakeProfit and
// TrailingStop.
type ModifyOrderArg interface {
	applyModifyOrderArg(url.Values)
}

func (u Units) applyModifyOrderArg(v url.Values) {
	optionalArgs(v).SetInt("units", int(u))
}

func (p Price) applyModifyOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("price", float64(p))
}

func (e Expiry) applyModifyOrderArg(v url.Values) {
	optionalArgs(v).SetTime("expiry", time.Time(e))
}

func (lb LowerBound) applyModifyOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("lowerBound", float64(lb))
}

func (ub UpperBound) applyModifyOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("upperBound", float64(ub))
}

func (sl StopLoss) applyModifyOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("stopLoss", float64(sl))
}

func (tp TakeProfit) applyModifyOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("takeProfit", float64(tp))
}

func (ts TrailingStop) applyModifyOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("trailingStop", float64(ts))
}

// ModifyOrder updates an open order. Supported arguments are Units(), Price(), Expiry(),
// UpperBound(), StopLoss(), TakeProfit() and TrailingStop().
func (c *Client) ModifyOrder(orderId int, arg ModifyOrderArg, args ...ModifyOrderArg) (*Order, error) {
	data := url.Values{}
	arg.applyModifyOrderArg(data)
	for _, arg = range args {
		arg.applyModifyOrderArg(data)
	}
	o := struct {
		ApiError
		Order
	}{}
	urlStr := fmt.Sprintf("/v1/accounts/%d/orders/%d", c.accountId, orderId)
	if err := requestAndDecode(c, "PATCH", urlStr, data, &o); err != nil {
		return nil, err
	}
	return &o.Order, nil
}

type CancelOrderResponse struct {
	TransactionId int       `json:"id"`
	Instrument    string    `json:"instrument"`
	Units         int       `json:"units"`
	Side          string    `json:"side"`
	Price         float64   `json:"price"`
	Time          time.Time `json:"time"`
}

// CancelOrder closes an open order.
func (c *Client) CancelOrder(orderId int) (*CancelOrderResponse, error) {
	urlStr := fmt.Sprintf("/v1/accounts/%d/orders/%d", c.accountId, orderId)
	cor := struct {
		ApiError
		CancelOrderResponse
	}{}
	if err := requestAndDecode(c, "DELETE", urlStr, nil, &cor); err != nil {
		return nil, err
	}
	return &cor.CancelOrderResponse, nil
}
