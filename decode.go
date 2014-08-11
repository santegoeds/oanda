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

func (dec *jsonDecoder) Decode(vp interface{}) error {
	value := reflect.ValueOf(vp)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return fmt.Errorf("Decode argument is not a pointer or is nil")
	}

	m := map[string]json.RawMessage{}
	err := dec.dec.Decode(&m)
	if err != nil {
		return err
	}

	if _, ok := m["code"]; ok {
		apiError := &ApiError{}
		if err = json.Unmarshal(m["code"], &apiError.Code); err != nil {
			return err
		}
		if apiError.Code != 0 {
			json.Unmarshal(m["message"], &apiError.Message)
			json.Unmarshal(m["moreinfo"], &apiError.MoreInfo)
			return apiError
		}
	}

	value = value.Elem()
	valueType := value.Type()

	switch value.Kind() {
	default:
		return fmt.Errorf("invalid type %s; only struct and maps are supported", value.Kind())

	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			fieldValue := value.Field(i)
			fieldType := valueType.Field(i)

			fieldName := fieldType.Tag.Get("json")
			if fieldName == "" {
				fieldName = fieldType.Name
			}

			jsonText, ok := m[fieldName]
			if !ok {
				continue
			}

			if err = json.Unmarshal(jsonText, fieldValue.Addr().Interface()); err != nil {
				return err
			}
		}

	case reflect.Map:
		dstMap := *(vp.(*map[string]interface{}))
		for k, v := range m {
			var dstValue interface{}
			if err = json.Unmarshal(v, &dstValue); err != nil {
				return err
			}
			dstMap[k] = dstValue
		}
	}

	return nil
}
