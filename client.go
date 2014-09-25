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
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var debug = false

var (
	DefaultDateFormat  = DateFormat("RFC3339")
	DefaultContentType = ContentType("application/x-www-form-urlencoded")
	DefaultTransport   = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,

		// The number of open connections to the stream server are restricted. Disable support for
		// idle connections.
		MaxIdleConnsPerHost: -1,
	}
)

///////////////////////////////////////////////////////////////////////////////////////////////////
// RequestModifiers

// A RequestModifier updates an http.Request before it is passed to an http.Client for execution.
type RequestModifier interface {
	Modify(*http.Request)
}

type TokenAuthenticator string

func (a TokenAuthenticator) Modify(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+string(a))
}

type UsernameAuthenticator string

func (a UsernameAuthenticator) Modify(req *http.Request) {
	u := req.URL
	q := u.Query()
	q.Set("username", string(a))
	u.RawQuery = q.Encode()
}

type Environment string

func (e Environment) Modify(req *http.Request) {
	u := req.URL
	envStr := string(e)
	if envStr == "sandbox" {
		u.Scheme = "http"
	} else {
		u.Scheme = "https"
	}
	if u.Host == "" {
		u.Host = "api-" + string(e) + ".oanda.com"
	}
}

type DateFormat string

func (d DateFormat) Modify(req *http.Request) {
	req.Header.Set("X-Accept-Datetime-Format", string(d))
}

type ContentType string

func (c ContentType) Modify(req *http.Request) {
	if req.Body != nil {
		req.Header.Set("Content-Type", string(c))
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Client

type Client struct {
	ReqMods   []RequestModifier
	accountId int
	*http.Client
}

func NewFxPracticeClient(token string) (*Client, error) {
	return newClient(Environment("fxpractice"), TokenAuthenticator(token)), nil
}

func NewFxTradeClient(token string) (*Client, error) {
	return newClient(Environment("fxtrade"), TokenAuthenticator(token)), nil
}

func NewSandboxClient() (*Client, error) {
	c := newClient(Environment("sandbox"))
	if userName, err := initSandboxAccount(c); err != nil {
		return nil, err
	} else {
		c.ReqMods = append(c.ReqMods, UsernameAuthenticator(userName))
	}
	return c, nil
}

// SelectAccount configures the account for which requests are executed.  AccountId 0 means that
// further requests are for all accounts.
func (c *Client) SelectAccount(accountId int) {
	c.accountId = accountId
}

func (c *Client) NewRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	for _, reqMod := range c.ReqMods {
		reqMod.Modify(req)
	}
	return req, nil
}

func (c *Client) CancelRequest(req *http.Request) {
	type canceler interface {
		CancelRequest(*http.Request)
	}
	tr, ok := c.Transport.(canceler)
	if ok {
		tr.CancelRequest(req)
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// PollRequest

type PollRequest struct {
	c   *Client
	req *http.Request
}

func (pr *PollRequest) Poll() (*http.Response, error) {
	rsp, err := pr.c.Do(pr.req)
	if err != nil {
		return nil, err
	}
	etag := rsp.Header.Get("ETag")
	if etag != "" {
		pr.req.Header.Set("If-None-Match", etag)
	}
	return rsp, nil
}

func newClient(reqMod ...RequestModifier) *Client {
	c := Client{
		ReqMods: []RequestModifier{
			DefaultDateFormat,
			DefaultContentType,
		},
		Client: &http.Client{
			Transport: DefaultTransport,
		},
	}
	c.ReqMods = append(c.ReqMods, reqMod...)
	return &c
}

// initSandboxAccount creates a new test account in the sandbox environment and adds a
// RequestModifier for authentication to the client.
func initSandboxAccount(c *Client) (string, error) {
	v := struct {
		ApiError
		Username  string `json:"username"`
		Password  string `json:"password"`
		AccountId int    `json:"accountId"`
	}{}
	if err := requestAndDecode(c, "POST", "/v1/accounts", nil, &v); err != nil {
		return "", err
	}
	return v.Username, nil
}

type ReturnCodeChecker interface {
	CheckReturnCode() error
}

// ApiError hold error details as returned by the Oanda servers.
type ApiError struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	MoreInfo string `json:"moreInfo"`
}

func (ae *ApiError) Error() string {
	return fmt.Sprintf("ApiError{Code: %d, Message: %s, Moreinfo: %s}",
		ae.Code, ae.Message, ae.MoreInfo)
}

func (ae *ApiError) CheckReturnCode() error {
	if ae.Code != 0 {
		return ae
	}
	return nil
}

func getAndDecode(c *Client, urlStr string, vp ReturnCodeChecker) error {
	return requestAndDecode(c, "GET", urlStr, nil, vp)
}

func requestAndDecode(c *Client, method, urlStr string, data url.Values, vp ReturnCodeChecker) error {
	var rdr io.Reader
	if len(data) > 0 {
		rdr = strings.NewReader(data.Encode())
	}
	req, err := c.NewRequest(method, urlStr, rdr)
	if err != nil {
		return err
	}
	rsp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	dec := json.NewDecoder(rsp.Body)
	if err = dec.Decode(vp); err != nil {
		return err
	}
	if err = vp.CheckReturnCode(); err != nil {
		return err
	}
	return nil
}
