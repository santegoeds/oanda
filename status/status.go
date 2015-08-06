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

// Package status provides functions to query the current and past status of the services that
// oanda makes available via the REST API.
//
// For further information see the Oanda documentation at http://api-status.oanda.com/documentation

package status

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

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

// ApiService represents information about a service.
type ApiService struct {
	Id           string           `json:"id"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	List         *ApiServiceList  `json:"list"`
	CurrentEvent *ApiServiceEvent `json:"current-event"`
	Url          string           `json:"url"`
}

// Services returns an array with information about all existing services.
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

// Service returns information about the service with the specified service id.
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

// Represents information about a service list. Services can be grouped together by linking
// multiple services with the same service list.
type ApiServiceList struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Url         string `json:"url"`
}

// ServiceLists returns an array with information off all defined service lists.
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

// ServiceList returns information about the service list with the specified service id.
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

// ApiServiceEvent represents an event that potentially changed the status of a service.
type ApiServiceEvent struct {
	Sid           string            `json:"sid"`
	Message       string            `json:"message"`
	Timestamp     string            `json:"timestamp"`
	Url           string            `json:"url"`
	Status        *ApiServiceStatus `json:"status"`
	Informational bool              `json:"informational"`
}

// ServiceEvents returns an array of events for the specified service id. If start- and/or end is
// not nil the list if filtered to include only the events between start- and end time, inclusive.
//
// Note that only the date part of the start- and end times considered and parts with finer
// granularity are ignored.
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

// CurrentServiceEvent returns event information for the current (i.e. most recent) event.
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

// ServiceEvent return information about the service event that matches the specified serviceId
// and eventId.
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

// ApServiceStatus represents the status of an oanda service.
type ApiServiceStatus struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Url         string `json:"url"`
	Level       string `json:"level"`
	Image       string `json:"image"`
	Default     bool   `json:"default"`
}

// ServiceStatuses returns an array with status information for each defined service.
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

// ServiceStatus return status information about the service with the specifed id.
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
// Status images

type ApiStatusImage struct {
	Name    string `json:"name"`
	IconSet string `json:"icon_set"`
	Url     string `json:"url"`
}

func StatusImages() ([]ApiStatusImage, error) {
	v := struct {
		ClientError
		Images []ApiStatusImage `json:"images"`
	}{}
	if err := getStatus("/v1/status-images", &v); err != nil {
		return nil, err
	}
	if v.IsError {
		return nil, &v.ClientError
	}
	return v.Images, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// private

func getStatus(urlStr string, v interface{}) error {
	urlStr = "http://api-status.oanda.com/api" + urlStr
	rsp, err := http.Get(urlStr)
	if err != nil {
		return err
	}
	defer func() {
		io.Copy(ioutil.Discard, rsp.Body)
		rsp.Body.Close()
	}()
	d := json.NewDecoder(rsp.Body)
	if err := d.Decode(v); err != nil {
		return err
	}
	return nil
}
