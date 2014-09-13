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
	"reflect"
)

// ApiError hold error details as returned by the Oanda servers.
type ApiError struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	MoreInfo string `json:"moreInfo"`
}

func (ae ApiError) Error() string {
	return fmt.Sprintf("ApiError{Code: %d, Message: %s, Moreinfo: %s}",
		ae.Code, ae.Message, ae.MoreInfo)
}

type jsonDecoder struct {
	dec *json.Decoder
}

// NewDecoder returns a new Decoder that checks for an Oanda Api error.
func NewDecoder(r io.Reader) *jsonDecoder {
	return &jsonDecoder{json.NewDecoder(r)}
}

func (dec *jsonDecoder) Decode(vp interface{}) (err error) {
	if err = dec.dec.Decode(vp); err != nil {
		return
	}

	value := reflect.ValueOf(vp).Elem()
	switch value.Kind() {
	case reflect.Struct:
		err = apiErrorFromStruct(value)
	case reflect.Map:
		err = apiErrorFromMap(vp)
	default:
		err = errors.New("Unsupported map value type.")
	}
	return
}

func apiErrorFromStruct(value reflect.Value) error {
	apiErr := value.FieldByName("ApiError")
	if !apiErr.IsValid() {
		return errors.New("struct does not embed an ApiError instance.")
	}
	if apiErr.Kind() != reflect.Struct {
		return errors.New("Embedded ApiError field is not of type oanda.ApiError")
	}
	codeField := apiErr.FieldByName("Code")
	if !codeField.IsValid() || codeField.Kind() != reflect.Int {
		return errors.New("Embedded ApiError field is not of type oanda.ApiError")
	}
	// Not an error
	if codeField.Int() == 0 {
		return nil
	}
	// Return the embedded ApiError field as the error
	return apiErr.Addr().Interface().(*ApiError)
}

func apiErrorFromMap(vp interface{}) error {
	if imPtr, ok := vp.(*map[string]interface{}); ok {
		if code, ok := (*imPtr)["code"]; ok {
			apiErr := ApiError{}
			if fcode, ok := code.(float64); !ok {
				return fmt.Errorf("unexpected code type %v", code)
			} else {
				apiErr.Code = int(fcode)
			}
			if str, ok := (*imPtr)["message"]; ok {
				if apiErr.Message, ok = str.(string); !ok {
					return fmt.Errorf("unexpected message type %v", str)
				}
			}
			if str, ok := (*imPtr)["moreInfo"]; ok {
				if apiErr.MoreInfo, ok = str.(string); !ok {
					return fmt.Errorf("unexpected moreInfo type %v", str)
				}
			}
			return &apiErr
		}
		return nil
	}

	if rmPtr, ok := vp.(*map[string]json.RawMessage); ok {
		if data, ok := (*rmPtr)["code"]; ok {
			apiErr := ApiError{}
			if err := json.Unmarshal(data, &apiErr.Code); err != nil {
				return err
			}
			if apiErr.Code != 0 {
				if err := json.Unmarshal((*rmPtr)["message"], &apiErr.Message); err != nil {
					return err
				}
				if err := json.Unmarshal((*rmPtr)["moreInfo"], &apiErr.MoreInfo); err != nil {
					return err
				}
			}
			return &apiErr
		}
		return nil
	}

	return errors.New("unsupported map type")
}
