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
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var debug = false

type Client struct {
	AccountId int

	transport  *http.Transport
	httpClient *http.Client
	token      string
	env        string
}

func NewFxPracticeClient(token string) *Client {
	return newClient("fxpractice", token)
}

func NewFxTradeClient(token string) *Client {
	return newClient("fxtrade", token)
}

type SandboxClient struct {
	*Client
}

func NewSandboxClient() *SandboxClient {
	c := &SandboxClient{
		Client: newClient("sandbox", ""),
	}
	return c
}

func newClient(env, token string) *Client {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	client := Client{
		transport: tr,
		httpClient: &http.Client{
			Transport: tr,
		},
		token: token,
		env:   env,
	}
	return &client
}

func (c *Client) CancelRequest(req *http.Request) {
	c.transport.CancelRequest(req)
}

type Context struct {
	client *Client
	req    *http.Request
}

func (c *Client) newContext(method string, u *url.URL, data url.Values) (*Context, error) {
	var rdr io.Reader
	if len(data) > 0 {
		rdr = strings.NewReader(data.Encode())
	}
	method = strings.ToUpper(method)
	req, err := http.NewRequest(method, u.String(), rdr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Accept-Datetime-Format", "RFC3339")
	if rdr != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if c.token != "" {
		if c.env == "sandbox" {
			q := req.URL.Query()
			q.Set("username", c.token)
			req.URL.RawQuery = q.Encode()
		} else {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}
	}

	return &Context{client: c, req: req}, nil
}

// Request initiates an Http connection using the request on the Context.
func (ctx *Context) Request() (*http.Response, error) {
	rsp, err := ctx.client.httpClient.Do(ctx.req)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func (ctx *Context) CancelRequest() {
	ctx.client.CancelRequest(ctx.req)
}

func (ctx *Context) Decode(vp interface{}) (int64, error) {
	rsp, err := ctx.Request()
	if err != nil {
		return 0, err
	}
	defer rsp.Body.Close()

	var rdr io.Reader = rsp.Body
	if debug {
		fmt.Fprintln(os.Stderr, rsp)
		rdr = io.TeeReader(rdr, os.Stderr)
	}

	if rsp.ContentLength != 0 {
		eTag := rsp.Header.Get("ETag")
		if eTag != "" {
			ctx.req.Header.Set("If-None-Match", eTag)
		}

		dec := NewDecoder(rdr)
		if err = dec.Decode(vp); err != nil {
			return rsp.ContentLength, err
		}
	}

	return rsp.ContentLength, nil
}

func (c *Client) getUrl(urlPath string, hostPrefix string) *url.URL {
	u := url.URL{
		Path: urlPath,
	}

	u.Host = fmt.Sprintf("%s-%s.oanda.com", hostPrefix, c.env)

	switch c.env {
	case "sandbox":
		u.Scheme = "http"
	default:
		u.Scheme = "https"
	}

	return &u
}
