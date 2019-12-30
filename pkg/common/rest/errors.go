/*
Copyright (c) 2019 Inspur, Inc.
All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License"); you may
not use this file except in compliance with the License. You may obtain
a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
License for the specific language governing permissions and limitations
under the License.
*/

package rest

import (
	"fmt"
)

const (
	RestFailureUnknown        = 1
	RestResourceBusy          = 2
	RestRequestMalfunction    = 3
	RestResourceDNE           = 4
	RestUnableToConnect       = 5
	RestRPM                   = 6 // Response Processing Malfunction
	RestStorageFailureUnknown = 7
	RestObjectExists          = 8
)

type RestError interface {
	Error() (out string)
	GetCode() int
}

type restError struct {
	msg  string
	code int
}

//TODO: Refactor to move logging of error message in this func
func GetError(c int, m string) RestError {
	out := restError{
		code: c,
		msg:  m,
	}
	return &out
}

func (err *restError) Error() (out string) {

	switch (*err).code {

	case RestResourceBusy:
		out = fmt.Sprintf("Resource is busy. %s", err.msg)

	case RestRequestMalfunction:
		out = fmt.Sprintf("Failure in sending data to storage: %s", err.msg)

	case RestRPM:
		out = fmt.Sprintf("Failure during processing response from storage: %s", err.msg)
	case RestResourceDNE:
		out = fmt.Sprintf("Resource %s do not exists", err.msg)
	case RestObjectExists:

		out = fmt.Sprintf("Object exists: %s", err.msg)

	case RestStorageFailureUnknown:
		out = fmt.Sprintf("Storage failes with unknown error: %s", err.msg)

	default:
		out = fmt.Sprint("Unknown internal Error. %s", err.msg)

	}
	return out
}

func (err *restError) GetCode() int {
	return err.code

}
