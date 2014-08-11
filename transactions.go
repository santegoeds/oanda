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

type Transaction struct {
	TranId    int
	AccountId int
	Time      time.Time
	Type      string
	attrs     tranAttrs
}

type TradeData struct {
	TradeId  int
	Units    int
	Pl       float64
	Interest float64
}

type (
	MinId        int
	Transactions []Transaction
	tranAttrs    map[string]interface{}
)

// UnmarshalJSON populates a transaction with information from a Json byte array.
func (tran *Transaction) UnmarshalJSON(data []byte) (err error) {
	if err = json.Unmarshal(data, &tran.attrs); err != nil {
		return
	}

	if tran.TranId, err = tran.AsInt("id"); err != nil {
		return
	}
	delete(tran.attrs, "id")

	if tran.AccountId, err = tran.AsInt("accountId"); err != nil {
		return
	}
	delete(tran.attrs, "accountId")

	if tran.Type, err = tran.AsString("type"); err != nil {
		return
	}
	delete(tran.attrs, "type")

	if tran.Time, err = tran.AsTime("time"); err != nil {
		return
	}
	delete(tran.attrs, "time")

	return nil
}

// String implementes the Stringer interface.
func (tran *Transaction) String() string {
	return fmt.Sprintf("Transaction{TranId: %d, Type: %s}", tran.TranId, tran.Type)
}

// AsFloat returns a Transaction attribute as a float value.
func (tran *Transaction) AsFloat(key string) (f float64, err error) {
	if v, ok := tran.attrs[key]; !ok {
		err = fmt.Errorf("invalid attribute %s", key)
		return

	} else if f, ok = v.(float64); !ok {
		err = fmt.Errorf("cannot convert attribute %s = %v to float64", key, v)
		return
	}
	return
}

// AsInt returns a Transction attribute as an integer value.
func (tran *Transaction) AsInt(key string) (i int, err error) {
	f, err := tran.AsFloat(key)
	if err != nil {
		return
	}

	if i = int(f); float64(i) != f {
		err = fmt.Errorf("cannot convert attribute %s = %v to integer", key, f)
		return
	}
	return
}

// AsString returns a Transaction attribute as a string value.
func (tran *Transaction) AsString(key string) (str string, err error) {
	if v, ok := tran.attrs[key]; !ok {
		err = fmt.Errorf("invalid attribute %s", key)
		return

	} else if str, ok = v.(string); !ok {
		err = fmt.Errorf("cannot convert attribute %s = %v to string", key, v)
		return
	}

	return
}

// AsTime returns a Transaction attribute as a time value.
func (tran *Transaction) AsTime(key string) (t time.Time, err error) {
	s, err := tran.AsString(key)
	if err != nil {
		return
	}
	t, err = time.Parse(time.RFC3339, s)
	return
}

// TradeData returns any details about a trade that was created or reduced as a result of
// the transaction.
func (tran *Transaction) TradeData() (td *TradeData, err error) {
	m, ok := tran.attrs["tradeOpened"].(map[string]interface{})
	if !ok {
		m, ok = tran.attrs["tradeReduced"].(map[string]interface{})
	}
	if !ok {
		err = fmt.Errorf("%s does not have TradeData", tran)
		return
	}

	td = &TradeData{}

	f, ok := m["id"].(float64)
	if !ok {
		err = fmt.Errorf("%s does not have TradeData", tran)
		return
	}
	td.TradeId = int(f)

	if f, ok = m["units"].(float64); !ok {
		err = fmt.Errorf("%s TradeData does not include units", tran)
		return
	}
	td.Units = int(f)

	if f, ok = m["pl"].(float64); ok {
		td.Pl = f
	}

	if f, ok = m["interest"].(float64); ok {
		td.Interest = f
	}

	return
}

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

	tran := Transaction{}
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
