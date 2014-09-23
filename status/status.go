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
package status

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type ApiServiceStatus struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Url         string `json:"url"`
	Level       string `json:"level"`
	Image       string `json:"image"`
	Default     bool   `json:"default"`
}

type ApiServiceList struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Url         string `json:"url"`
}

type ApiServiceEvent struct {
	Sid           string            `json:"sid"`
	Message       string            `json:"message"`
	Timestamp     string            `json:"timestamp"`
	Url           string            `json:"url"`
	Status        *ApiServiceStatus `json:"status"`
	Informational bool              `json:"informational"`
}

type ApiService struct {
	Id           string           `json:"id"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	List         *ApiServiceList  `json:"list"`
	CurrentEvent *ApiServiceEvent `json:"current-event"`
	Url          string           `json:"url"`
}

type ClientError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	IsError bool   `json:"error"`
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("ClientError{Code: %d, Message: %s, IsError: %v}", e.Code, e.Message, e.IsError)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Service

func Services() ([]ApiService, error) {
	v := struct {
		ClientError
		Services []ApiService `json:"services"`
	}{}
	if err := getStatus("/v1/services", &v); err != nil {
		return nil, err
	}
	if v.IsError {
		return nil, &v.ClientError
	}
	return v.Services, nil
}

func Service(serviceId string) (*ApiService, error) {
	v := struct {
		ClientError
		ApiService
	}{}
	if err := getStatus(fmt.Sprintf("/v1/services/%s", serviceId), &v); err != nil {
		return nil, err
	}
	if v.IsError {
		return nil, &v.ClientError
	}
	return &v.ApiService, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Service List

func ServiceLists() ([]ApiServiceList, error) {
	v := struct {
		ClientError
		Lists []ApiServiceList `json:"lists"`
	}{}
	if err := getStatus("/v1/service-lists", &v); err != nil {
		return nil, err
	}
	if v.IsError {
		return nil, &v.ClientError
	}
	return v.Lists, nil
}

func ServiceList(serviceId string) (*ApiServiceList, error) {
	v := struct {
		ClientError
		ApiServiceList
	}{}
	if err := getStatus(fmt.Sprintf("/v1/service-lists/%s", serviceId), &v); err != nil {
		return nil, err
	}
	if v.IsError {
		return nil, &v.ClientError
	}
	return &v.ApiServiceList, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Events

func ServiceEvents(serviceId string, start *time.Time, end *time.Time) ([]ApiServiceEvent, error) {
	v := struct {
		ClientError
		Events []ApiServiceEvent `json:"events"`
	}{}
	u, err := url.Parse(fmt.Sprintf("/v1/services/%s/events", serviceId))
	if err != nil {
		return nil, err
	}
	q := u.Query()
	if start != nil {
		q.Set("start", start.Truncate(24*time.Hour).Format(time.RFC1123))
	}
	if end != nil {
		q.Set("end", end.Truncate(24*time.Hour).Format(time.RFC1123))
	}
	u.RawQuery = q.Encode()
	if err = getStatus(u.String(), &v); err != nil {
		return nil, err
	}
	if v.IsError {
		return nil, &v.ClientError
	}
	return v.Events, nil
}

func CurrentServiceEvent(serviceId string) (*ApiServiceEvent, error) {
	v := struct {
		Code    int  `json:"code"`
		IsError bool `json:"error"`
		ApiServiceEvent
	}{}
	if err := getStatus(fmt.Sprintf("/v1/services/%s/events/current", serviceId), &v); err != nil {
		return nil, err
	}
	if v.IsError {
		clientError := ClientError{
			Code:    v.Code,
			IsError: v.IsError,
			Message: v.Message,
		}
		return nil, &clientError
	}
	return &v.ApiServiceEvent, nil
}

func ServiceEvent(serviceId, eventId string) (*ApiServiceEvent, error) {
	v := struct {
		Code    int  `json:"code"`
		IsError bool `json:"error"`
		ApiServiceEvent
	}{}
	if err := getStatus(fmt.Sprintf("/v1/services/%s/events/%s", serviceId, eventId), &v); err != nil {
		return nil, err
	}
	if v.IsError {
		clientError := ClientError{
			Code:    v.Code,
			IsError: v.IsError,
			Message: v.Message,
		}
		return nil, &clientError
	}
	return &v.ApiServiceEvent, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// Status

func ServiceStatuses() ([]ApiServiceStatus, error) {
	v := struct {
		ClientError
		Statuses []ApiServiceStatus `json:"statuses"`
	}{}
	if err := getStatus("/v1/statuses", &v); err != nil {
		return nil, err
	}
	if v.IsError {
		return nil, &v.ClientError
	}
	return v.Statuses, nil
}

func ServiceStatus(statusId string) (*ApiServiceStatus, error) {
	v := struct {
		ClientError
		ApiServiceStatus
	}{}
	if err := getStatus(fmt.Sprintf("/v1/statuses/%s", statusId), &v); err != nil {
		return nil, err
	}
	if v.IsError {
		return nil, &v.ClientError
	}
	return &v.ApiServiceStatus, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// private
func getStatus(urlStr string, v interface{}) error {
	urlStr = "http://api-status.oanda.com/api" + urlStr
	rsp, err := http.Get(urlStr)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	d := json.NewDecoder(rsp.Body)
	if err := d.Decode(v); err != nil {
		return err
	}
	return nil
}
