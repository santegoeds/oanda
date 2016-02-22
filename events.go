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
)

///////////////////////////////////////////////////////////////////////////////////////////////////
// Events

type evtTradeDetailData struct {
	TradeId  int     `json:"id"`
	Units    int     `json:"units"`
	Pl       float64 `json:"pl"`
	Interest float64 `json:"interest"`
}

type evtTradeDetail struct{ content *evtTradeDetailData }

func (td *evtTradeDetail) TradeId() int      { return td.content.TradeId }
func (td *evtTradeDetail) Units() int        { return td.content.Units }
func (td *evtTradeDetail) Pl() float64       { return td.content.Pl }
func (td *evtTradeDetail) Interest() float64 { return td.content.Interest }

type evtHeaderContent struct {
	TranId    uint64 `json:"id"`
	AccountId int    `json:"accountId"`
	Time      Time   `json:"time"`
	Type      string `json:"type"`
}

type evtHeader struct {
	content *evtHeaderContent
}

type evtBody struct {
	Instrument               string              `json:"instrument"`
	Side                     string              `json:"side"`
	Units                    int                 `json:"units"`
	Price                    float64             `json:"price"`
	Expiry                   Time                `json:"expiry"`
	Reason                   string              `json:"reason"`
	LowerBound               float64             `json:"lowerBound"`
	UpperBound               float64             `json:"upperBound"`
	TakeProfitPrice          float64             `json:"takeProfitPrice"`
	StopLossPrice            float64             `json:"stopLossPrice"`
	TrailingStopLossDistance float64             `json:"trailingStopLossDistance"`
	Pl                       float64             `json:"pl"`
	Interest                 float64             `json:"interest"`
	AccountBalance           float64             `json:"accountBalance"`
	Rate                     float64             `json:"rate"`
	Amount                   float64             `json:"amount"`
	TradeId                  int                 `json:"tradeId"`
	OrderId                  int                 `json:"orderId"`
	TradeOpened              *evtTradeDetailData `json:"tradeOpened"`
	TradeReduced             *evtTradeDetailData `json:"tradeReduced"`
	HomeCurrency             string              `json:"homeCurrency"`
}

type Event interface {
	TranId() uint64
	AccountId() int
	Time() Time
	Type() string
}

func (t *evtHeader) TranId() uint64 { return t.content.TranId }
func (t *evtHeader) AccountId() int { return t.content.AccountId }
func (t *evtHeader) Time() Time     { return t.content.Time }
func (t *evtHeader) Type() string   { return t.content.Type }

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *evtHeader) UnmarshalJSON(data []byte) (err error) {
	return json.Unmarshal(data, &t.content)
}

// String implementes the fmt.Stringer interface.
func (t evtHeader) String() string {
	return fmt.Sprintf("Event{TranId: %d, AccountId: %d Type: %s}",
		t.TranId(), t.AccountId(), t.Type())
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// CREATE

// AccountCreateEvent represents an event of type CREATE.
type AccountCreateEvent struct {
	evtHeader
	body *evtBody
}

func (t *AccountCreateEvent) HomeCurrency() string { return t.body.HomeCurrency }
func (t *AccountCreateEvent) Reason() string       { return t.body.Reason }

///////////////////////////////////////////////////////////////////////////////////////////////////
// MARKET_ORDER_CREATE

// TradeCreateEvent represents an event of type MARKET_ORDER_CREATE
type TradeCreateEvent struct {
	evtHeader
	body *evtBody
}

func (t *TradeCreateEvent) Instrument() string       { return t.body.Instrument }
func (t *TradeCreateEvent) Side() string             { return t.body.Side }
func (t *TradeCreateEvent) Units() int               { return t.body.Units }
func (t *TradeCreateEvent) Price() float64           { return t.body.Price }
func (t *TradeCreateEvent) Pl() float64              { return t.body.Pl }
func (t *TradeCreateEvent) Interest() float64        { return t.body.Interest }
func (t *TradeCreateEvent) LowerBound() float64      { return t.body.LowerBound }
func (t *TradeCreateEvent) UpperBound() float64      { return t.body.UpperBound }
func (t *TradeCreateEvent) AccountBalance() float64  { return t.body.AccountBalance }
func (t *TradeCreateEvent) StopLossPrice() float64   { return t.body.StopLossPrice }
func (t *TradeCreateEvent) TakeProfitPrice() float64 { return t.body.TakeProfitPrice }
func (t *TradeCreateEvent) TrailingStopLossDistance() float64 {
	return t.body.TrailingStopLossDistance
}
func (t *TradeCreateEvent) TradeOpened() *evtTradeDetail {
	if t.body.TradeOpened != nil {
		return &evtTradeDetail{t.body.TradeOpened}
	}
	return nil
}
func (t *TradeCreateEvent) TradeReduced() *evtTradeDetail {
	if t.body.TradeReduced != nil {
		return &evtTradeDetail{t.body.TradeReduced}
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// LIMIT_ORDER_CREATE, STOP_ORDER_CREATE, MARKET_IF_TOUCHED_CREATE

// OrderCreateEvent represents an event of type LIMIT_ORDER_CREATE, STOP_ORDER_CREATE or
// MARKET_IF_TOUCHED_CREATE
type OrderCreateEvent struct {
	evtHeader
	body *evtBody
}

func (t *OrderCreateEvent) Instrument() string       { return t.body.Instrument }
func (t *OrderCreateEvent) Side() string             { return t.body.Side }
func (t *OrderCreateEvent) Units() int               { return t.body.Units }
func (t *OrderCreateEvent) Price() float64           { return t.body.Price }
func (t *OrderCreateEvent) Expiry() Time             { return t.body.Expiry }
func (t *OrderCreateEvent) Reason() string           { return t.body.Reason }
func (t *OrderCreateEvent) LowerBound() float64      { return t.body.LowerBound }
func (t *OrderCreateEvent) UpperBound() float64      { return t.body.UpperBound }
func (t *OrderCreateEvent) TakeProfitPrice() float64 { return t.body.TakeProfitPrice }
func (t *OrderCreateEvent) StopLossPrice() float64   { return t.body.StopLossPrice }
func (t *OrderCreateEvent) TrailingStopLossDistance() float64 {
	return t.body.TrailingStopLossDistance
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// ORDER_UPDATE

// OrderUpdateEvent represents an event of type ORDER_UPDATE
type OrderUpdateEvent struct {
	evtHeader
	body *evtBody
}

func (t *OrderUpdateEvent) Instrument() string       { return t.body.Instrument }
func (t *OrderUpdateEvent) Side() string             { return t.body.Side }
func (t *OrderUpdateEvent) Units() int               { return t.body.Units }
func (t *OrderUpdateEvent) Reason() string           { return t.body.Reason }
func (t *OrderUpdateEvent) LowerBound() float64      { return t.body.LowerBound }
func (t *OrderUpdateEvent) UpperBound() float64      { return t.body.UpperBound }
func (t *OrderUpdateEvent) TakeProfitPrice() float64 { return t.body.TakeProfitPrice }
func (t *OrderUpdateEvent) StopLossPrice() float64   { return t.body.StopLossPrice }
func (t *OrderUpdateEvent) TrailingStopLossDistance() float64 {
	return t.body.TrailingStopLossDistance
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// ORDER_CANCEL

// OrderCancelEvent represents and event of type ORDER_CANCEL.
type OrderCancelEvent struct {
	evtHeader
	body *evtBody
}

func (t *OrderCancelEvent) OrderId() int   { return t.body.OrderId }
func (t *OrderCancelEvent) Reason() string { return t.body.Reason }

///////////////////////////////////////////////////////////////////////////////////////////////////
// ORDER_FILLED

// OrderFilledEvent represents an event of type ORDER_FILLED.
type OrderFilledEvent struct {
	evtHeader
	body *evtBody
}

func (t *OrderFilledEvent) OrderId() int { return t.body.OrderId }

///////////////////////////////////////////////////////////////////////////////////////////////////
// TRADE_UPDATE

// TradeUpdateEvent represents an event of type TRADE_UPDATE
type TradeUpdateEvent struct {
	evtHeader
	body *evtBody
}

func (t *TradeUpdateEvent) Instrument() string               { return t.body.Instrument }
func (t *TradeUpdateEvent) Units() int                       { return t.body.Units }
func (t *TradeUpdateEvent) Side() string                     { return t.body.Side }
func (t *TradeUpdateEvent) TradeId() int                     { return t.body.TradeId }
func (t *TradeUpdateEvent) TakeProfitPrice() float64         { return t.body.TakeProfitPrice }
func (t *TradeUpdateEvent) StopLossPrice() float64           { return t.body.StopLossPrice }
func (t *TradeUpdateEvent) TailingStopLossDistance() float64 { return t.body.TrailingStopLossDistance }

///////////////////////////////////////////////////////////////////////////////////////////////////
// TRADE_CLOSE, MIGRATE_TRADE_CLOSE, TAKE_PROFIT_FILLED, STOP_LOSS_FILLED, TRAILING_STOP_FILLED,
// MARGIN_CLOSEOUT

// TradeCloseEvent represents an event of type TRADE_CLOSE, MIGRATE_TRADE_CLOSE, TAKE_PROFIT_FILLED,
// STOP_LOSS_FILLED, TRAILING_STOP_FILLED or MARGIN_CLOSEOUT.
type TradeCloseEvent struct {
	evtHeader
	body *evtBody
}

func (t *TradeCloseEvent) Instrument() string      { return t.body.Instrument }
func (t *TradeCloseEvent) Units() int              { return t.body.Units }
func (t *TradeCloseEvent) Side() string            { return t.body.Side }
func (t *TradeCloseEvent) Price() float64          { return t.body.Price }
func (t *TradeCloseEvent) Pl() float64             { return t.body.Pl }
func (t *TradeCloseEvent) Interest() float64       { return t.body.Interest }
func (t *TradeCloseEvent) AccountBalance() float64 { return t.body.AccountBalance }
func (t *TradeCloseEvent) TradeId() int            { return t.body.TradeId }

///////////////////////////////////////////////////////////////////////////////////////////////////
// MIGRATE_TRADE_OPEN

// MigrateTradeOpenEvent represents an event of type MIGRATE_TRADE_OPEN
type MigrateTradeOpenEvent struct {
	evtHeader
	body *evtBody
}

func (t *MigrateTradeOpenEvent) Instrument() string       { return t.body.Instrument }
func (t *MigrateTradeOpenEvent) Side() string             { return t.body.Side }
func (t *MigrateTradeOpenEvent) Units() int               { return t.body.Units }
func (t *MigrateTradeOpenEvent) Price() float64           { return t.body.Price }
func (t *MigrateTradeOpenEvent) TakeProfitPrice() float64 { return t.body.TakeProfitPrice }
func (t *MigrateTradeOpenEvent) StopLossPrice() float64   { return t.body.StopLossPrice }
func (t *MigrateTradeOpenEvent) TrailingStopLossDistance() float64 {
	return t.body.TrailingStopLossDistance
}
func (t *MigrateTradeOpenEvent) TradeOpened() *evtTradeDetail {
	return &evtTradeDetail{t.body.TradeOpened}
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// SET_MARGIN_RATE

// SetMarginRateEvent represent and event of type SET_MARGIN_RATE.
type SetMarginRateEvent struct {
	evtHeader
	body *evtBody
}

func (t *SetMarginRateEvent) Rate() float64 { return t.body.Rate }

///////////////////////////////////////////////////////////////////////////////////////////////////
// TRANSFER_FUNDS

// TransferFundsEvent represents an event of type TRANSFER_FUNDS.
type TransferFundsEvent struct {
	evtHeader
	body *evtBody
}

func (t *TransferFundsEvent) Amount() float64 { return t.body.Amount }

///////////////////////////////////////////////////////////////////////////////////////////////////
// DAILY_INTEREST

// DailyInterestEvent represents an event of type DAILY_INTEREST.
type DailyInterestEvent struct {
	evtHeader
	body *evtBody
}

func (t *DailyInterestEvent) Interest() float64 { return t.body.Interest }

///////////////////////////////////////////////////////////////////////////////////////////////////
// FEE

// FeeEvent represents an event of type FEE
type FeeEvent struct {
	evtHeader
	body *evtBody
}

func (t *FeeEvent) Amount() float64         { return t.body.Amount }
func (t *FeeEvent) AccountBalance() float64 { return t.body.AccountBalance }
func (t *FeeEvent) Reason() string          { return t.body.Reason }

type (
	MinId int
)

// EventsArg is an optional argument for method PollEvents.
type EventsArg interface {
	applyEventsArg(url.Values)
}

func (mi MaxId) applyEventsArg(v url.Values) {
	optionalArgs(v).SetInt("maxId", int(mi))
}

func (mi MinId) applyEventsArg(v url.Values) {
	optionalArgs(v).SetInt("minId", int(mi))
}

func (c Count) applyEventsArg(v url.Values) {
	optionalArgs(v).SetInt("count", int(c))
}

func (i Instrument) applyEventsArg(v url.Values) {
	v.Set("instrument", string(i))
}

func (ids Ids) applyEventsArg(v url.Values) {
	optionalArgs(v).SetIntArray("ids", []int(ids))
}

// PollEvents returns an array of events. Optional arguments are MaxId, MinId, Count,
// Instrument and Ids.
//
// See http://developer.oanda.com/docs/v1/transactions/#get-transaction-history for further
// information.
func (c *Client) PollEvents(args ...EventsArg) ([]Event, error) {
	urlStr := fmt.Sprintf("/v1/accounts/%d/transactions", c.accountId)
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	data := u.Query()
	for _, arg := range args {
		arg.applyEventsArg(data)
	}
	u.RawQuery = data.Encode()
	urlStr = u.String()

	s := struct {
		Events []struct {
			*evtHeaderContent
			*evtBody
		} `json:"transactions"`
	}{}
	if err = getAndDecode(c, urlStr, &s); err != nil {
		return nil, err
	}
	events := []Event{}
	for _, rawEvent := range s.Events {
		evt, err := asEvent(rawEvent.evtHeaderContent, rawEvent.evtBody)
		if err != nil {
			return nil, err
		}
		events = append(events, evt)
	}
	return events, nil
}

// PollEvent returns data for a single event.
func (c *Client) PollEvent(tranId uint64) (Event, error) {
	evtData := struct {
		evtHeaderContent
		evtBody
	}{}
	urlStr := fmt.Sprintf("/v1/accounts/%d/transactions/%d", c.accountId, tranId)
	if err := getAndDecode(c, urlStr, &evtData); err != nil {
		return nil, err
	}
	return asEvent(&evtData.evtHeaderContent, &evtData.evtBody)
}

func asEvent(header *evtHeaderContent, body *evtBody) (Event, error) {
	switch header.Type {
	case "CREATE":
		return &AccountCreateEvent{evtHeader{header}, body}, nil
	case "MARKET_ORDER_CREATE":
		return &TradeCloseEvent{evtHeader{header}, body}, nil
	case "LIMIT_ORDER_CREATE", "STOP_ORDER_CREATE", "MARKET_IF_TOUCHED_CREATE":
		return &OrderCreateEvent{evtHeader{header}, body}, nil
	case "ORDER_UPDATE":
		return &OrderUpdateEvent{evtHeader{header}, body}, nil
	case "ORDER_CANCEL":
		return &OrderCancelEvent{evtHeader{header}, body}, nil
	case "ORDER_FILLED":
		return &OrderFilledEvent{evtHeader{header}, body}, nil
	case "TRADE_UPDATE":
		return &TradeUpdateEvent{evtHeader{header}, body}, nil
	case "TRADE_CLOSE", "MIGRATE_TRADE_CLOSE", "TAKE_PROFIT_FILLED", "STOP_LOSS_FILLED", "TRAILING_STOP_FILLED",
		"MARGIN_CLOSEOUT":
		return &TradeCloseEvent{evtHeader{header}, body}, nil
	case "MIGRATE_TRADE_OPEN":
		return &MigrateTradeOpenEvent{evtHeader{header}, body}, nil
	case "SET_MARGIN_RATE":
		return &SetMarginRateEvent{evtHeader{header}, body}, nil
	case "TRANSFER_FUNDS":
		return &TransferFundsEvent{evtHeader{header}, body}, nil
	case "DAILY_INTEREST":
		return &DailyInterestEvent{evtHeader{header}, body}, nil
	case "FEE":
		return &FeeEvent{evtHeader{header}, body}, nil
	}
	return nil, fmt.Errorf("Unexpected event type %s", header.Type)
}

// FullEventHistory returns a url from which a file containing the full transaction history
// for the account can be downloaded.
func (c *Client) FullEventHistory() (*url.URL, error) {
	urlStr := fmt.Sprintf("/v1/accounts/%d/alltransactions", c.accountId)
	req, err := c.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	rsp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	// FIXME: Return the io.ReadCloser to the data instead of the location URL.  Might want to
	// wrap that in a streamServer wrapper so that the request can be interrupted?
	tranUrl, err := rsp.Location()
	if err != nil {
		return nil, err
	}
	return tranUrl, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// EventServer

// An EventServer receives events (aka transactions) for one or more accountId(s).
type EventServer struct {
	// If HeartbeatFunc is not nil it is invoked once for every heartbeat message that the
	// EventServer receives.
	HeartbeatFunc HeartbeatHandlerFunc
	chanMap       *eventChans
	srv           *messageServer
}

type (
	EventHandlerFunc func(int, Event)
)

// NewEventServer returns an server instance for receiving events for the specified accountId(s).
// If no accountId is specified events for all accountIds are received.  Note that the sandbox
// environment requires that at least one accountId is provided.
//
// See http://developer.oanda.com/docs/v1/stream/#events-streaming for further information.
func (c *Client) NewEventServer(accountId ...int) (*EventServer, error) {
	req, err := c.NewRequest("GET", "/v1/events", nil)
	if err != nil {
		return nil, err
	}
	useStreamHost(req)

	q := req.URL.Query()
	optionalArgs(q).SetIntArray("accountIds", accountId)
	req.URL.RawQuery = q.Encode()

	es := &EventServer{
		chanMap: newEventChans(accountId),
	}

	streamSrv := StreamServer{
		handleMessagesFn:   es.handleMessages,
		handleHeartbeatsFn: es.handleHeartbeats,
	}

	if s, err := c.newMessageServer(req, streamSrv); err != nil {
		return nil, err
	} else {
		es.srv = s
	}

	return es, nil
}

// ConnectAndDispatch starts the event server and blocks until Stop() is called.  Function handleFn
// is called for each event that is received.
//
// See http://developer.oanda.com/docs/v1/stream/ and http://developer.oanda.com/docs/v1/transactions/
// for further information.
func (es *EventServer) ConnectAndHandle(handleFn EventHandlerFunc) (err error) {
	es.initServer(handleFn)
	return es.srv.ConnectAndDispatch()
}

// Stop terminates the events server and causes ConnectAndHandle() to return.
func (es *EventServer) Stop() {
	es.srv.Stop()
}

func (es *EventServer) initServer(handleFn EventHandlerFunc) {
	for _, accId := range es.chanMap.AccountIds() {
		evtC := make(chan Event, defaultBufferSize)
		es.chanMap.Set(accId, evtC)

		go func(lclC <-chan Event) {
			for evt := range lclC {
				handleFn(evt.AccountId(), evt)
			}
		}(evtC)
	}
	return
}

func (es *EventServer) handleHeartbeats(hbC <-chan Time) {
	for hb := range hbC {
		if es.HeartbeatFunc != nil {
			es.HeartbeatFunc(hb)
		}
	}
}

func (es *EventServer) handleMessages(msgC <-chan StreamMessage) {
	for msg := range msgC {
		// FIXME: id is already a maxed out int 32, for type:SET_MARGIN_RATE.
		fmt.Println(msg)
		// StreamMessage{transaction, {"id":4294967295,"accountId":XXXX,"time":"1456149472000000","type":"SET_MARGIN_RATE","marginRate":0.05}}

		rawEvent := struct {
			*evtHeaderContent
			*evtBody
		}{}
		if err := json.Unmarshal(msg.RawMessage, &rawEvent); err != nil {
			// FIXME: log message
			return
		}
		evt, err := asEvent(rawEvent.evtHeaderContent, rawEvent.evtBody)
		if err != nil {
			// FIXME: Log error
			return
		}
		evtC, ok := es.chanMap.Get(evt.AccountId())
		if !ok {
			// FIXME: log error "unexpected accountId"
		} else if evtC != nil {
			evtC <- evt
		} else {
			// FiXME: log "event after server closed"
		}
	}

	for _, accId := range es.chanMap.AccountIds() {
		evtC, _ := es.chanMap.Get(accId)
		es.chanMap.Set(accId, nil)
		if evtC != nil {
			close(evtC)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// private

type eventChans struct {
	mtx sync.RWMutex
	m   map[int]chan Event
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

func (ec *eventChans) Set(accountId int, ch chan Event) {
	ec.mtx.Lock()
	defer ec.mtx.Unlock()
	ec.m[accountId] = ch
}

func (ec *eventChans) Get(accountId int) (chan Event, bool) {
	ec.mtx.RLock()
	defer ec.mtx.RUnlock()
	ch, ok := ec.m[accountId]
	return ch, ok
}

func newEventChans(accountIds []int) *eventChans {
	m := make(map[int]chan Event, len(accountIds))
	for _, accId := range accountIds {
		m[accId] = nil
	}
	return &eventChans{
		m: m,
	}
}
