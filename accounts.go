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
)

// Account represents an Oanda account.
type Account struct {
	AccountId       int      `json:"accountId"`
	Name            string   `json:"accountName"`
	Balance         float64  `json:"balance"`
	UnrealizedPl    float64  `json:"unrealizedPl"`
	RealizedPl      float64  `json:"realizedPl"`
	MarginUsed      float64  `json:"marginUsed"`
	MarginAvailable float64  `json:"marginAvailable"`
	OpenTrades      int      `json:"openTrades"`
	OpenOrders      int      `json:"openOrders"`
	Currency        string   `json:"accountCurrency"`
	MarginRate      float64  `json:"marginRate"`
	PropertyName    []string `json:"accountPropertyName"`
}

// String implements the Stringer interface.
func (a Account) String() string {
	return fmt.Sprintf("Account{AccountId: %d, Name: %s, Currency: %s}", a.AccountId, a.Name,
		a.Currency)
}

type Accounts []Account

// Accounts returns a list with all the know accounts.
func (c *Client) Accounts() ([]Account, error) {
	ctx, err := c.newContext("GET", c.getUrl("/v1/accounts", "api"), nil)
	if err != nil {
		return nil, err
	}

	v := struct {
		ApiError
		Accounts []Account `json:"accounts"`
	}{}

	if _, err = ctx.Decode(&v); err != nil {
		return nil, err
	}
	return v.Accounts, nil
}

// NewAccount queries the Oanda servers for account information for the specified accountId
// and returns a new Account instance.
func (c *Client) Account(accountId int) (*Account, error) {
	ctx, err := c.newContext("GET", c.getUrl(fmt.Sprintf("/v1/accounts/%d", accountId), "api"), nil)
	if err != nil {
		return nil, err
	}

	acc := struct {
		ApiError
		Account
	}{}
	if _, err = ctx.Decode(&acc); err != nil {
		return nil, err
	}
	return &acc.Account, nil
}

// CreateAccount creates a new test account in the Oanda Sandbox environment.
func (c *SandboxClient) NewAccount() (int, error) {
	ctx, err := c.newContext("POST", c.getUrl("/v1/accounts", "api"), nil)
	if err != nil {
		return 0, err
	}

	v := struct {
		ApiError
		Username  string `json:"username"`
		Password  string `json:"password"`
		AccountId int    `json:"accountId"`
	}{}

	if _, err := ctx.Decode(&v); err != nil {
		return 0, err
	}
	c.Client.token = v.Username
	return v.AccountId, nil
}
