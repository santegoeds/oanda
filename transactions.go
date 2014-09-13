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
	"time"
)

///////////////////////////////////////////////////////////////////////////////////////////////////
// Unified transaction

type tranTradeDetailData struct {
	TradeId  int     `json:"id"`
	Units    int     `json:"units"`
	Pl       float64 `json:"pl"`
	Interest float64 `json:"interest"`
}

type tranTradeDetail struct{ data *tranTradeDetailData }

func (td *tranTradeDetail) TradeId() int      { return td.data.TradeId }
func (td *tranTradeDetail) Units() int        { return td.data.Units }
func (td *tranTradeDetail) Pl() float64       { return td.data.Pl }
func (td *tranTradeDetail) Interest() float64 { return td.data.Interest }

type tranData struct {
	TranId                   int                  `json:"id"`
	AccountId                int                  `json:"accountId"`
	Time                     time.Time            `json:"time"`
	Type                     string               `json:"type"`
	Instrument               string               `json:"instrument"`
	Side                     string               `json:"side"`
	Units                    int                  `json:"units"`
	Price                    float64              `json:"price"`
	Expiry                   time.Time            `json:"expiry"`
	Reason                   string               `json:"reason"`
	LowerBound               float64              `json:"lowerBound"`
	UpperBound               float64              `json:"upperBound"`
	TakeProfitPrice          float64              `json:"takeProfitPrice"`
	StopLossPrice            float64              `json:"stopLossPrice"`
	TrailingStopLossDistance float64              `json:"trailingStopLossDistance"`
	Pl                       float64              `json:"pl"`
	Interest                 float64              `json:"interest"`
	AccountBalance           float64              `json:"accountBalance"`
	Rate                     float64              `json:"rate"`
	Amount                   float64              `json:"amount"`
	TradeId                  int                  `json:"tradeId"`
	OrderId                  int                  `json:"orderId"`
	TradeOpened              *tranTradeDetailData `json:"tradeOpened"`
	TradeReduced             *tranTradeDetailData `json:"tradeReduced"`
	HomeCurrency             string               `json:"homeCurrency"`
}

type transaction struct{ data tranData }

// TranId returns the transaction id for the transaction.
func (t *transaction) TranId() int { return t.data.TranId }

// AccountId returns the account id where the transaction occurred.
func (t *transaction) AccountId() int { return t.data.AccountId }

// Time returns the time at which the transaction occurred.
func (t *transaction) Time() time.Time { return t.data.Time }

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *transaction) UnmarshalJSON(data []byte) (err error) {
	return json.Unmarshal(data, &t.data)
}

// String implementes the Stringer interface.
func (t *transaction) String() string {
	return fmt.Sprintf("transaction{TranId: %d, AccountId: %d Type: %s}",
		t.TranId(), t.AccountId(), t.Type())
}

// Type returns the transaction type.
func (t *transaction) Type() string { return t.data.Type }

func (t *transaction) AsAccountCreate() (*tranAccountCreate, error) {
	if t.Type() == "CREATE" {
		return &tranAccountCreate{t}, nil
	}
	return nil, fmt.Errorf("conversion AsAccountCreate is invalid for type %s", t.Type())
}

func (t *transaction) AsTradeCreate() (*tranTradeCreate, error) {
	switch t.Type() {
	case "MARKET_ORDER_CREATE", "ORDER_FILLED":
		return &tranTradeCreate{t}, nil
	}
	return nil, fmt.Errorf("conversion AsTradeCreate is invalid for type %s", t.Type())
}

// AsOrderCreate returns an OrderCreate transaction instance to provide access to
// transaction specific information.  An error is returned if the transaction is not of
// type LIMIT_ORDER_CREATE, STOP_ORDER_CREATE or MARKET_IF_TOUCHED_ORDER_CREATE.
func (t *transaction) AsOrderCreate() (*tranOrderCreate, error) {
	switch t.Type() {
	case "LIMIT_ORDER_CREATE", "STOP_ORDER_CREATE", "MARKET_IF_TOUCHED_ORDER_CREATE":
		return &tranOrderCreate{t}, nil
	}
	return nil, fmt.Errorf("conversion AsOrderCreate is invalid for type %s", t.Type())
}

func (t *transaction) AsOrderUpdate() (*tranOrderUpdate, error) {
	if t.Type() == "ORDER_UPDATE" {
		return &tranOrderUpdate{t}, nil
	}
	return nil, fmt.Errorf("conversion AsOrderUpdate is invalid for type %s", t.Type())
}

func (t *transaction) AsOrderCancel() (*tranOrderCancel, error) {
	if t.Type() == "ORDER_CANCEL" {
		return &tranOrderCancel{t}, nil
	}
	return nil, fmt.Errorf("conversion AsOrderCancel is invalid for type %s", t.Type())
}

func (t *transaction) AsOrderFilled() (*tranOrderFilled, error) {
	if t.Type() == "ORDER_FILLED" {
		return &tranOrderFilled{&tranTradeCreate{t}}, nil
	}
	return nil, fmt.Errorf("conversion AsOrderFilled is invalid for type %s", t.Type())
}

func (t *transaction) AsTradeUpdate() (*tranTradeUpdate, error) {
	if t.Type() == "TRADE_UPDATE" {
		return &tranTradeUpdate{t}, nil
	}
	return nil, fmt.Errorf("conversion AsTradeUpdate is invalid for type %s", t.Type())
}

func (t *transaction) AsTradeClose() (*tranTradeClose, error) {
	switch t.Type() {
	case "TRADE_CLOSE", "MIGRATE_TRADE_CLOSE",
		"TAKE_PROFIT_FILLED", "STOP_LOSS_FILLED",
		"TRAILING_STOP_FILLED", "MARGIN_CLOSEOUT":

		return &tranTradeClose{t}, nil
	}
	return nil, fmt.Errorf("conversion AsTradeClose is invalid for type %s", t.Type())
}

func (t *transaction) AsMigrateTradeOpen() (*tranMigrateTradeOpen, error) {
	if t.Type() == "MIGRATE_TRADE_OPEN" {
		return &tranMigrateTradeOpen{t}, nil
	}
	return nil, fmt.Errorf("conversion AsMigrateTradeOpen is invalid for type %s", t.Type())
}

func (t *transaction) AsSetMarginRate() (*tranSetMarginRate, error) {
	if t.Type() == "SET_MARGIN_RATE" {
		return &tranSetMarginRate{t}, nil
	}
	return nil, fmt.Errorf("conversion AsSetMarginRate is invalid for type %s", t.Type())
}

func (t *transaction) AsTransferFunds() (*tranTransferFunds, error) {
	if t.Type() == "TRANSFER_FUNDS" {
		return &tranTransferFunds{t}, nil
	}
	return nil, fmt.Errorf("conversion AsTransferFunds is invalid for type %s", t.Type())
}

func (t *transaction) AsDailyInterest() (*tranDailyInterest, error) {
	if t.Type() == "DAILY_INTEREST" {
		return &tranDailyInterest{t}, nil
	}
	return nil, fmt.Errorf("conversion AsDailyInterest is invalid for type %s", t.Type())
}

func (t *transaction) AsFee() (*tranFee, error) {
	if t.Type() == "FEE" {
		return &tranFee{t}, nil
	}
	return nil, fmt.Errorf("conversion AsFee is invalid for type %s", t.Type())
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// CREATE

type tranAccountCreate struct{ *transaction }

func (t *tranAccountCreate) HomeCurrency() string { return t.data.HomeCurrency }
func (t *tranAccountCreate) Reason() string       { return t.data.Reason }

///////////////////////////////////////////////////////////////////////////////////////////////////
// MARKET_ORDER_CREATE

type tranTradeCreate struct{ *transaction }

func (t *tranTradeCreate) Instrument() string       { return t.data.Instrument }
func (t *tranTradeCreate) Side() string             { return t.data.Side }
func (t *tranTradeCreate) Units() int               { return t.data.Units }
func (t *tranTradeCreate) Price() float64           { return t.data.Price }
func (t *tranTradeCreate) Pl() float64              { return t.data.Pl }
func (t *tranTradeCreate) Interest() float64        { return t.data.Interest }
func (t *tranTradeCreate) LowerBound() float64      { return t.data.LowerBound }
func (t *tranTradeCreate) UpperBound() float64      { return t.data.UpperBound }
func (t *tranTradeCreate) AccountBalance() float64  { return t.data.AccountBalance }
func (t *tranTradeCreate) StopLossPrice() float64   { return t.data.StopLossPrice }
func (t *tranTradeCreate) TakeProfitPrice() float64 { return t.data.TakeProfitPrice }
func (t *tranTradeCreate) TrailingStopLossDistance() float64 {
	return t.data.TrailingStopLossDistance
}
func (t *tranTradeCreate) TradeOpened() *tranTradeDetail {
	if t.data.TradeOpened != nil {
		return &tranTradeDetail{t.data.TradeOpened}
	}
	return nil
}
func (t *tranTradeCreate) TradeReduced() *tranTradeDetail {
	if t.data.TradeReduced != nil {
		return &tranTradeDetail{t.data.TradeReduced}
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// LIMIT_ORDER_CREATE, STOP_ORDER_CREATE, MARKET_IF_TOUCHED_CREATE

type tranOrderCreate struct{ *transaction }

func (t *tranOrderCreate) Instrument() string       { return t.data.Instrument }
func (t *tranOrderCreate) Side() string             { return t.data.Side }
func (t *tranOrderCreate) Units() int               { return t.data.Units }
func (t *tranOrderCreate) Price() float64           { return t.data.Price }
func (t *tranOrderCreate) Expiry() time.Time        { return t.data.Expiry }
func (t *tranOrderCreate) Reason() string           { return t.data.Reason }
func (t *tranOrderCreate) LowerBound() float64      { return t.data.LowerBound }
func (t *tranOrderCreate) UpperBound() float64      { return t.data.UpperBound }
func (t *tranOrderCreate) TakeProfitPrice() float64 { return t.data.TakeProfitPrice }
func (t *tranOrderCreate) StopLossPrice() float64   { return t.data.StopLossPrice }
func (t *tranOrderCreate) TrailingStopLossDistance() float64 {
	return t.data.TrailingStopLossDistance
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// ORDER_UPDATE

type tranOrderUpdate struct{ *transaction }

func (t *tranOrderUpdate) Instrument() string       { return t.data.Instrument }
func (t *tranOrderUpdate) Side() string             { return t.data.Side }
func (t *tranOrderUpdate) Units() int               { return t.data.Units }
func (t *tranOrderUpdate) Reason() string           { return t.data.Reason }
func (t *tranOrderUpdate) LowerBound() float64      { return t.data.LowerBound }
func (t *tranOrderUpdate) UpperBound() float64      { return t.data.UpperBound }
func (t *tranOrderUpdate) TakeProfitPrice() float64 { return t.data.TakeProfitPrice }
func (t *tranOrderUpdate) StopLossPrice() float64   { return t.data.StopLossPrice }
func (t *tranOrderUpdate) TrailingStopLossDistance() float64 {
	return t.data.TrailingStopLossDistance
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// ORDER_CANCEL

type tranOrderCancel struct{ *transaction }

func (t *tranOrderCancel) OrderId() int   { return t.data.OrderId }
func (t *tranOrderCancel) Reason() string { return t.data.Reason }

///////////////////////////////////////////////////////////////////////////////////////////////////
// ORDER_FILLED

type tranOrderFilled struct{ *tranTradeCreate }

func (t *tranOrderFilled) OrderId() int { return t.data.OrderId }

///////////////////////////////////////////////////////////////////////////////////////////////////
// TRADE_UPDATE

type tranTradeUpdate struct{ *transaction }

func (t *tranTradeUpdate) Instrument() string               { return t.data.Instrument }
func (t *tranTradeUpdate) Units() int                       { return t.data.Units }
func (t *tranTradeUpdate) Side() string                     { return t.data.Side }
func (t *tranTradeUpdate) TradeId() int                     { return t.data.TradeId }
func (t *tranTradeUpdate) TakeProfitPrice() float64         { return t.data.TakeProfitPrice }
func (t *tranTradeUpdate) StopLossPrice() float64           { return t.data.StopLossPrice }
func (t *tranTradeUpdate) TailingStopLossDistance() float64 { return t.data.TrailingStopLossDistance }

///////////////////////////////////////////////////////////////////////////////////////////////////
// TRADE_CLOSE, MIGRATE_TRADE_CLOSE, TAKE_PROFIT_FILLED, STOP_LOSS_FILLED, TRAILING_STOP_FILLED,
// MARGIN_CLOSEOUT

type tranTradeClose struct{ *transaction }

func (t *tranTradeClose) Instrument() string {
	return t.data.Instrument
}

func (t *tranTradeClose) Units() int {
	return t.data.Units
}

func (t *tranTradeClose) Side() string {
	return t.data.Side
}

func (t *tranTradeClose) Price() float64 {
	return t.data.Price
}

func (t *tranTradeClose) Pl() float64 {
	return t.data.Pl
}

func (t *tranTradeClose) Interest() float64 {
	return t.data.Interest
}

func (t *tranTradeClose) AccountBalance() float64 {
	return t.data.AccountBalance
}

func (t *tranTradeClose) TradeId() int {
	return t.data.TradeId
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// MIGRATE_TRADE_OPEN

type tranMigrateTradeOpen struct{ *transaction }

func (t *tranMigrateTradeOpen) Instrument() string       { return t.data.Instrument }
func (t *tranMigrateTradeOpen) Side() string             { return t.data.Side }
func (t *tranMigrateTradeOpen) Units() int               { return t.data.Units }
func (t *tranMigrateTradeOpen) Price() float64           { return t.data.Price }
func (t *tranMigrateTradeOpen) TakeProfitPrice() float64 { return t.data.TakeProfitPrice }
func (t *tranMigrateTradeOpen) StopLossPrice() float64   { return t.data.StopLossPrice }
func (t *tranMigrateTradeOpen) TrailingStopLossDistance() float64 {
	return t.data.TrailingStopLossDistance
}
func (t *tranMigrateTradeOpen) TradeOpened() *tranTradeDetail {
	return &tranTradeDetail{t.data.TradeOpened}
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// SET_MARGIN_RATE

type tranSetMarginRate struct{ *transaction }

func (t *tranSetMarginRate) Rate() float64 { return t.data.Rate }

///////////////////////////////////////////////////////////////////////////////////////////////////
// TRANSFER_FUNDS

type tranTransferFunds struct{ *transaction }

func (t *tranTransferFunds) Amount() float64 { return t.data.Amount }

///////////////////////////////////////////////////////////////////////////////////////////////////
// DAILY_INTEREST

type tranDailyInterest struct{ *transaction }

func (t *tranDailyInterest) Interest() float64 { return t.data.Interest }

///////////////////////////////////////////////////////////////////////////////////////////////////
// FEE

type tranFee struct{ *transaction }

func (t *tranFee) Amount() float64         { return t.data.Amount }
func (t *tranFee) AccountBalance() float64 { return t.data.AccountBalance }
func (t *tranFee) Reason() string          { return t.data.Reason }

type (
	MinId        int
	transactions []transaction
)

type TransactionsArg interface {
	ApplyTransactionsArg(url.Values)
}

func (mi MaxId) ApplyTransactionsArg(v url.Values) {
	optionalArgs(v).SetInt("maxId", int(mi))
}

func (mi MinId) ApplyTransactionsArg(v url.Values) {
	optionalArgs(v).SetInt("minId", int(mi))
}

func (c Count) ApplyTransactionsArg(v url.Values) {
	optionalArgs(v).SetInt("count", int(c))
}

func (i Instrument) ApplyTransactionsArg(v url.Values) {
	v.Set("instrument", string(i))
}

func (ids Ids) ApplyTransactionsArg(v url.Values) {
	optionalArgs(v).SetIntArray("ids", []int(ids))
}

// Transactions returns an array of transactions.
func (c *Client) Transactions(args ...TransactionsArg) (transactions, error) {
	u := c.getUrl(fmt.Sprintf("/v1/accounts/%d/transactions", c.AccountId), "api")

	data := u.Query()
	for _, arg := range args {
		arg.ApplyTransactionsArg(data)
	}
	u.RawQuery = data.Encode()

	ctx, err := c.newContext("GET", u, nil)
	if err != nil {
		return nil, err
	}

	s := struct {
		Transactions transactions `json:"transactions"`
	}{}
	if _, err := ctx.Decode(&s); err != nil {
		return nil, err
	}
	return s.Transactions, nil
}

// Transaction returns data for a single transaction.
func (c *Client) Transaction(tranId int) (*transaction, error) {
	u := c.getUrl(fmt.Sprintf("/v1/accounts/%d/transactions/%d", c.AccountId, tranId), "api")
	ctx, err := c.newContext("GET", u, nil)
	if err != nil {
		return nil, err
	}

	tran := transaction{}
	if _, err = ctx.Decode(&tran); err != nil {
		return nil, err
	}

	return &tran, nil
}

// FullTransactionHistory returns a url from which a file containing the full transaction history
// for the account can be downloaded.
func (c *Client) FullTransactionHistory() (*url.URL, error) {
	u := c.getUrl(fmt.Sprintf("/v1/accounts/%d/alltransactions", c.AccountId), "api")
	ctx, err := c.newContext("GET", u, nil)
	if err != nil {
		return nil, err
	}

	rsp, err := ctx.Connect()
	if err != nil {
		return nil, err
	}

	tranUrl, err := url.Parse(rsp.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	return tranUrl, nil
}
