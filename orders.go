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
	Ts_Buy  TradeSide = "buy"
	Ts_Sell TradeSide = "sell"

	Ot_MarketIfTouched OrderType = "marketIfTouched"
	Ot_Limit           OrderType = "limit"
	Ot_Stop            OrderType = "stop"
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

// String implements the Stringer interface.
func (o Order) String() string {
	return fmt.Sprintf("Order{OrderId: %d, Side: %s, Units: %d, Instrument: %s}", o.OrderId, o.Side,
		o.Units, o.Instrument)
}

type (
	Orders []Order

	LowerBound   float64
	UpperBound   float64
	StopLoss     float64
	TakeProfit   float64
	TrailingStop float64
)

type NewOrderArg interface {
	ApplyNewOrderArg(url.Values)
}

func (lb LowerBound) ApplyNewOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("lowerBound", float64(lb))
}

func (ub UpperBound) ApplyNewOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("upperBound", float64(ub))
}

func (sl StopLoss) ApplyNewOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("stopLoss", float64(sl))
}

func (tp TakeProfit) ApplyNewOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("takeProfit", float64(tp))
}

func (ts TrailingStop) ApplyNewOrderArg(v url.Values) {
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
		arg.ApplyNewOrderArg(data)
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

// Order returns information for an existing order.
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

type (
	MaxId      int
	Count      int
	Instrument string
)

type OrdersArg interface {
	ApplyOrdersArg(url.Values)
}

func (mi MaxId) ApplyOrdersArg(v url.Values) {
	optionalArgs(v).SetInt("maxId", int(mi))
}

func (cnt Count) ApplyOrdersArg(v url.Values) {
	optionalArgs(v).SetInt("count", int(cnt))
}

func (in Instrument) ApplyOrdersArg(v url.Values) {
	v.Set("instrument", string(in))
}

// Orders returns an array with all orders that match the optional arguments (if any).
func (c *Client) Orders(args ...OrdersArg) (Orders, error) {
	u, err := url.Parse(fmt.Sprintf("/v1/accounts/%d/orders", c.accountId))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	for _, arg := range args {
		arg.ApplyOrdersArg(q)
	}
	u.RawQuery = q.Encode()

	rsp := struct {
		ApiError
		Orders Orders `json:"orders"`
	}{}
	if err := getAndDecode(c, u.String(), &rsp); err != nil {
		return nil, err
	}
	return rsp.Orders, nil
}

type (
	Units  int
	Expiry time.Time
	Price  float64
)

type ModifyOrderArg interface {
	ApplyModifyOrderArg(url.Values)
}

func (u Units) ApplyModifyOrderArg(v url.Values) {
	optionalArgs(v).SetInt("units", int(u))
}

func (p Price) ApplyModifyOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("price", float64(p))
}

func (e Expiry) ApplyModifyOrderArg(v url.Values) {
	optionalArgs(v).SetTime("expiry", time.Time(e))
}

func (lb LowerBound) ApplyModifyOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("lowerBound", float64(lb))
}

func (ub UpperBound) ApplyModifyOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("upperBound", float64(ub))
}

func (sl StopLoss) ApplyModifyOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("stopLoss", float64(sl))
}

func (tp TakeProfit) ApplyModifyOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("takeProfit", float64(tp))
}

func (ts TrailingStop) ApplyModifyOrderArg(v url.Values) {
	optionalArgs(v).SetFloat("trailingStop", float64(ts))
}

// ModifyOrder updates an open order.
func (c *Client) ModifyOrder(orderId int, arg ModifyOrderArg, args ...ModifyOrderArg) (*Order, error) {
	data := url.Values{}
	arg.ApplyModifyOrderArg(data)
	for _, arg = range args {
		arg.ApplyModifyOrderArg(data)
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
