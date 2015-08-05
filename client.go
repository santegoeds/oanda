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
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var debug = "" //"trace"

const (
	defaultDateFormat  = DateFormat("RFC3339")
	defaultContentType = ContentType("application/x-www-form-urlencoded")
)

var (
	DefaultHttpClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,

			// The number of open connections to the stream server are restricted. Disable support for
			// idle connections.
			MaxIdleConnsPerHost: -1,
		},
	}
)

///////////////////////////////////////////////////////////////////////////////////////////////////
// RequestModifiers

// A requestModifier updates an http.Request before it is passed onto an http.Client for execution.
type requestModifier interface {
	modify(*http.Request)
}

// A TokenAuthenticator adds a Bearer authentication to a request header.
type TokenAuthenticator string

func (a TokenAuthenticator) modify(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+string(a))
}

// A UsernameAuthenticator adds a username to a request header.
type UsernameAuthenticator string

func (a UsernameAuthenticator) modify(req *http.Request) {
	u := req.URL
	q := u.Query()
	q.Set("username", string(a))
	u.RawQuery = q.Encode()
}

type Environment string

func (e Environment) modify(req *http.Request) {
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

// A DateFormat add the desired DateTime format to a request header.
type DateFormat string

func (d DateFormat) modify(req *http.Request) {
	req.Header.Set("X-Accept-Datetime-Format", string(d))
}

// A ContentType adds a Content-Type entry to the request header.
type ContentType string

func (c ContentType) modify(req *http.Request) {
	if req.Body != nil {
		req.Header.Set("Content-Type", string(c))
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Client

type Client struct {
	reqMods   []requestModifier
	accountId int
	*http.Client
}

func (c *Client) AccountId() int { return c.accountId }

// NewFxPracticeClient returns a client instance that connects to Oanda's fxpractice environment. String
// token should be set to the generated personal access token.
//
// See http://developer.oanda.com/docs/v1/auth/ for further information.
func NewFxPracticeClient(token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("No FxPractice access token")
	}
	return NewClient("fxpractice", token, nil)
}

// NewFxTradeClient returns a client instance that connects to Oanda's fxtrade environment. String token
// should be set to the generated personal access token.
//
// See http://developer.oanda.com/docs/v1/auth/ for further information.
func NewFxTradeClient(token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("No FxTrade access token")
	}
	return NewClient("fxtrade", token, nil)
}

// NewSandboxClient returns a client instance that connects to Oanda's fxsandbox environment. Creating a
// client will create a user in the sandbox environment with wich all further calls with be authenticated.
//
// See http://developer.oanda.com/docs/v1/auth/ for further information.
func NewSandboxClient() (*Client, error) {
	return NewClient("sandbox", "", nil)
}

func NewClient(environment string, token string, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = DefaultHttpClient
	}

	switch environment {
	case "fxpractice":
		return newClient(httpClient, Environment("fxpractice"), TokenAuthenticator(token)), nil
	case "fxtrade":
		return newClient(httpClient, Environment("fxtrade"), TokenAuthenticator(token)), nil
	case "sandbox":
		c := newClient(httpClient, Environment("sandbox"))
		if userName, err := initSandboxAccount(c); err != nil {
			return nil, err
		} else {
			c.reqMods = append(c.reqMods, UsernameAuthenticator(userName))
		}
		return c, nil
	}

	return nil, fmt.Errorf("Invalid Oanda environment %v", environment)
}

// SelectAccount configures an Oanda account.  All trades and orders will be booked under the
// selected account.   Use AccountId 0 to disable account selection.
func (c *Client) SelectAccount(accountId int) {
	c.accountId = accountId
}

// NewRequest creates a new http request.
func (c *Client) NewRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	for _, reqMod := range c.reqMods {
		reqMod.modify(req)
	}
	return req, nil
}

// CancelRequest aborts an in-progress HTTP request.
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

// PollRequest represents an http request that can be executed repeatedly.
type PollRequest struct {
	c   *Client
	req *http.Request
}

// Poll repeats the http request with which PollRequest was created.
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

func newClient(httpClient *http.Client, reqMod ...requestModifier) *Client {
	c := Client{
		reqMods: []requestModifier{
			defaultDateFormat,
			defaultContentType,
		},
		Client: httpClient,
	}
	c.reqMods = append(c.reqMods, reqMod...)
	return &c
}

// initSandboxAccount creates a new test account in the sandbox environment.
func initSandboxAccount(c *Client) (string, error) {
	v := struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		AccountId int    `json:"accountId"`
	}{}
	if err := requestAndDecode(c, "POST", "/v1/accounts", nil, &v); err != nil {
		return "", err
	}
	return v.Username, nil
}

type returnCodeChecker interface {
	checkReturnCode() error
}

// ApiError holds error details as returned by the Oanda servers.
type ApiError struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	MoreInfo string `json:"moreInfo"`
}

func (ae *ApiError) Error() string {
	return fmt.Sprintf("ApiError{Code: %d, Message: %s, Moreinfo: %s}",
		ae.Code, ae.Message, ae.MoreInfo)
}

func getAndDecode(c *Client, urlStr string, v interface{}) error {
	return requestAndDecode(c, "GET", urlStr, nil, v)
}

func requestAndDecode(c *Client, method, urlStr string, data url.Values, v interface{}) error {
	var rdr io.Reader
	if len(data) > 0 {
		rdr = strings.NewReader(data.Encode())
	}
	req, err := c.NewRequest(method, urlStr, rdr)
	if err != nil {
		return err
	}

	if debug == "trace" {
		fmt.Fprintln(os.Stderr, req)
		fmt.Fprintln(os.Stderr, data)
	}

	rsp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	var body io.Reader = rsp.Body
	if debug == "trace" {
		fmt.Println(os.Stderr, rsp)
		body = io.TeeReader(body, os.Stderr)
	}

	dec := json.NewDecoder(body)
	if rsp.StatusCode < 400 {
		return dec.Decode(v)
	}

	apiErr := ApiError{}
	if err = dec.Decode(&apiErr); err != nil {
		return err
	}
	return &apiErr
}
