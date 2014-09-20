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
	"sync"
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

type Transaction struct{ data tranData }

// TranId returns the transaction id for the transaction.
func (t *Transaction) TranId() int { return t.data.TranId }

// AccountId returns the account id where the transaction occurred.
func (t *Transaction) AccountId() int { return t.data.AccountId }

// Time returns the time at which the transaction occurred.
func (t *Transaction) Time() time.Time { return t.data.Time }

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Transaction) UnmarshalJSON(data []byte) (err error) {
	return json.Unmarshal(data, &t.data)
}

// String implementes the Stringer interface.
func (t *Transaction) String() string {
	return fmt.Sprintf("transaction{TranId: %d, AccountId: %d Type: %s}",
		t.TranId(), t.AccountId(), t.Type())
}

// Type returns the transaction type.
func (t *Transaction) Type() string { return t.data.Type }

func (t *Transaction) AsAccountCreate() (*tranAccountCreate, bool) {
	if t.Type() == "CREATE" {
		return &tranAccountCreate{t}, true
	}
	return nil, false
}

func (t *Transaction) AsTradeCreate() (*tranTradeCreate, bool) {
	switch t.Type() {
	case "MARKET_ORDER_CREATE", "ORDER_FILLED":
		return &tranTradeCreate{t}, true
	}
	return nil, false
}

// AsOrderCreate returns an OrderCreate transaction instance to provide access to
// transaction specific information.  An error is returned if the transaction is not of
// type LIMIT_ORDER_CREATE, STOP_ORDER_CREATE or MARKET_IF_TOUCHED_ORDER_CREATE.
func (t *Transaction) AsOrderCreate() (*tranOrderCreate, bool) {
	switch t.Type() {
	case "LIMIT_ORDER_CREATE", "STOP_ORDER_CREATE", "MARKET_IF_TOUCHED_ORDER_CREATE":
		return &tranOrderCreate{t}, true
	}
	return nil, false
}

func (t *Transaction) AsOrderUpdate() (*tranOrderUpdate, bool) {
	if t.Type() == "ORDER_UPDATE" {
		return &tranOrderUpdate{t}, true
	}
	return nil, false
}

func (t *Transaction) AsOrderCancel() (*tranOrderCancel, bool) {
	if t.Type() == "ORDER_CANCEL" {
		return &tranOrderCancel{t}, true
	}
	return nil, false
}

func (t *Transaction) AsOrderFilled() (*tranOrderFilled, bool) {
	if t.Type() == "ORDER_FILLED" {
		return &tranOrderFilled{&tranTradeCreate{t}}, true
	}
	return nil, false
}

func (t *Transaction) AsTradeUpdate() (*tranTradeUpdate, bool) {
	if t.Type() == "TRADE_UPDATE" {
		return &tranTradeUpdate{t}, true
	}
	return nil, false
}

func (t *Transaction) AsTradeClose() (*tranTradeClose, bool) {
	switch t.Type() {
	case "TRADE_CLOSE", "MIGRATE_TRADE_CLOSE",
		"TAKE_PROFIT_FILLED", "STOP_LOSS_FILLED",
		"TRAILING_STOP_FILLED", "MARGIN_CLOSEOUT":

		return &tranTradeClose{t}, true
	}
	return nil, false
}

func (t *Transaction) AsMigrateTradeOpen() (*tranMigrateTradeOpen, bool) {
	if t.Type() == "MIGRATE_TRADE_OPEN" {
		return &tranMigrateTradeOpen{t}, true
	}
	return nil, false
}

func (t *Transaction) AsSetMarginRate() (*tranSetMarginRate, bool) {
	if t.Type() == "SET_MARGIN_RATE" {
		return &tranSetMarginRate{t}, true
	}
	return nil, false
}

func (t *Transaction) AsTransferFunds() (*tranTransferFunds, bool) {
	if t.Type() == "TRANSFER_FUNDS" {
		return &tranTransferFunds{t}, true
	}
	return nil, false
}

func (t *Transaction) AsDailyInterest() (*tranDailyInterest, bool) {
	if t.Type() == "DAILY_INTEREST" {
		return &tranDailyInterest{t}, true
	}
	return nil, false
}

func (t *Transaction) AsFee() (*tranFee, bool) {
	if t.Type() == "FEE" {
		return &tranFee{t}, true
	}
	return nil, false
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// CREATE

type tranAccountCreate struct{ *Transaction }

func (t *tranAccountCreate) HomeCurrency() string { return t.data.HomeCurrency }
func (t *tranAccountCreate) Reason() string       { return t.data.Reason }

///////////////////////////////////////////////////////////////////////////////////////////////////
// MARKET_ORDER_CREATE

type tranTradeCreate struct{ *Transaction }

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

type tranOrderCreate struct{ *Transaction }

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

type tranOrderUpdate struct{ *Transaction }

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

type tranOrderCancel struct{ *Transaction }

func (t *tranOrderCancel) OrderId() int   { return t.data.OrderId }
func (t *tranOrderCancel) Reason() string { return t.data.Reason }

///////////////////////////////////////////////////////////////////////////////////////////////////
// ORDER_FILLED

type tranOrderFilled struct{ *tranTradeCreate }

func (t *tranOrderFilled) OrderId() int { return t.data.OrderId }

///////////////////////////////////////////////////////////////////////////////////////////////////
// TRADE_UPDATE

type tranTradeUpdate struct{ *Transaction }

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

type tranTradeClose struct{ *Transaction }

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

type tranMigrateTradeOpen struct{ *Transaction }

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

type tranSetMarginRate struct{ *Transaction }

func (t *tranSetMarginRate) Rate() float64 { return t.data.Rate }

///////////////////////////////////////////////////////////////////////////////////////////////////
// TRANSFER_FUNDS

type tranTransferFunds struct{ *Transaction }

func (t *tranTransferFunds) Amount() float64 { return t.data.Amount }

///////////////////////////////////////////////////////////////////////////////////////////////////
// DAILY_INTEREST

type tranDailyInterest struct{ *Transaction }

func (t *tranDailyInterest) Interest() float64 { return t.data.Interest }

///////////////////////////////////////////////////////////////////////////////////////////////////
// FEE

type tranFee struct{ *Transaction }

func (t *tranFee) Amount() float64         { return t.data.Amount }
func (t *tranFee) AccountBalance() float64 { return t.data.AccountBalance }
func (t *tranFee) Reason() string          { return t.data.Reason }

type (
	MinId        int
	Transactions []Transaction
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
func (c *Client) Transactions(args ...TransactionsArg) (Transactions, error) {
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
		ApiError
		Transactions Transactions `json:"transactions"`
	}{}
	if _, err := ctx.Decode(&s); err != nil {
		return nil, err
	}
	return s.Transactions, nil
}

// Transaction returns data for a single transaction.
func (c *Client) Transaction(tranId int) (*Transaction, error) {
	u := c.getUrl(fmt.Sprintf("/v1/accounts/%d/transactions/%d", c.AccountId, tranId), "api")
	ctx, err := c.newContext("GET", u, nil)
	if err != nil {
		return nil, err
	}

	tran := struct {
		ApiError
		Transaction
	}{}
	if _, err = ctx.Decode(&tran); err != nil {
		return nil, err
	}

	return &tran.Transaction, nil
}

// FullTransactionHistory returns a url from which a file containing the full transaction history
// for the account can be downloaded.
func (c *Client) FullTransactionHistory() (*url.URL, error) {
	u := c.getUrl(fmt.Sprintf("/v1/accounts/%d/alltransactions", c.AccountId), "api")
	ctx, err := c.newContext("GET", u, nil)
	if err != nil {
		return nil, err
	}

	if err := ctx.Request(); err != nil {
		return nil, err
	}

	tranUrl, err := ctx.Response().Location()
	if err != nil {
		return nil, err
	}

	return tranUrl, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// eventsServer

type (
	EventsHandlerFunc func(int, *Transaction)
	AccountId         int
)

type eventChans struct {
	mtx sync.RWMutex
	m   map[int]chan *Transaction
}

func (ec *eventChans) AccountIds() []int {
	ec.mtx.RLock()
	defer ec.mtx.RUnlock()
	accIds := make([]int, len(ec.m))
	for accId := range ec.m {
		accIds = append(accIds, accId)
	}
	return accIds
}

func (ec *eventChans) Set(accountId int, ch chan *Transaction) {
	ec.mtx.Lock()
	defer ec.mtx.Unlock()
	ec.m[accountId] = ch
}

func (ec *eventChans) Get(accountId int) (chan *Transaction, bool) {
	ec.mtx.RLock()
	defer ec.mtx.RUnlock()
	ch, ok := ec.m[accountId]
	return ch, ok
}

func newEventChans(accountIds []int) *eventChans {
	m := make(map[int]chan *Transaction, len(accountIds))
	for _, accId := range accountIds {
		m[accId] = nil
	}
	return &eventChans{
		m: m,
	}
}

type eventsServer struct {
	HeartbeatFunc HeartbeatHandlerFunc
	chanMap       *eventChans
	srv           *server
}

func (c *Client) NewEventsServer(accountId ...int) (*eventsServer, error) {
	u := c.getUrl("/v1/events", "stream")
	q := u.Query()
	optionalArgs(q).SetIntArray("accountIds", accountId)
	u.RawQuery = q.Encode()

	es := &eventsServer{
		chanMap: newEventChans(accountId),
	}

	streamSrv := StreamServer{
		HandleMessageFn:   es.handleMessage,
		HandleHeartbeatFn: es.handleHeartbeat,
	}

	if s, err := c.NewServer(u, streamSrv); err != nil {
		return nil, err
	} else {
		es.srv = s
	}

	return es, nil
}

func (es *eventsServer) Run(handleFn EventsHandlerFunc) (err error) {
	es.initServer(handleFn)
	defer es.finish()
	es.srv.Run()
	return err
}

func (es *eventsServer) Stop() {
	es.srv.Stop()
}

func (es *eventsServer) initServer(handleFn EventsHandlerFunc) {
	for _, accId := range es.chanMap.AccountIds() {
		tranC := make(chan *Transaction, defaultBufferSize)
		es.chanMap.Set(accId, tranC)

		go func(lclC <-chan *Transaction) {
			for tran := range lclC {
				handleFn(tran.AccountId(), tran)
			}
		}(tranC)
	}
	return
}

func (es *eventsServer) finish() {
	for _, accId := range es.chanMap.AccountIds() {
		tranC, _ := es.chanMap.Get(accId)
		es.chanMap.Set(accId, nil)
		if tranC != nil {
			close(tranC)
		}
	}
}

func (es *eventsServer) handleHeartbeat(hb time.Time) {
	if es.HeartbeatFunc != nil {
		es.HeartbeatFunc(hb)
	}
}

func (es *eventsServer) handleMessage(msgType string, rawMessage json.RawMessage) {
	tran := &Transaction{}
	if err := json.Unmarshal(rawMessage, tran); err != nil {
		// FIXME: log message
		return
	}
	tranC, ok := es.chanMap.Get(tran.AccountId())
	if !ok {
		// FIXME: log error "unexpected accountId"
	} else if tranC != nil {
		tranC <- tran
	} else {
		// FiXME: log "event after server closed"
	}
}
