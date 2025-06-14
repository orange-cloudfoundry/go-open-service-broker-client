/*
Copyright 2019 Orange Cloudfoundry.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v2

import (
	"fmt"
	"net/http"

	"k8s.io/klog/v2"
)

// internal message body types

type rotateBindingRequestBody struct {
	PredecessorBindingId *string `json:"predecessor_binding_id"`
}

func (c *client) RotateBinding(r *RotateBindingRequest) (*BindResponse, error) {
	if err := validateRotateBindingRequest(r); err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf(bindingURLFmt, c.URL, r.InstanceID, r.BindingID)
	params := map[string]string{}
	if r.AcceptsIncomplete {
		params[AcceptsIncomplete] = "true"
	}

	requestBody := &rotateBindingRequestBody{
		PredecessorBindingId: &r.PredecessorBindingID,
	}

	response, err := c.prepareAndDo(http.MethodPut, fullURL, params, requestBody, r.OriginatingIdentity)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = drainReader(response.Body)
		response.Body.Close()
	}()

	switch response.StatusCode {
	case http.StatusOK, http.StatusCreated:
		userResponse := &BindResponse{}
		if err := c.unmarshalResponse(response, userResponse); err != nil {
			return nil, HTTPStatusCodeError{StatusCode: response.StatusCode, ResponseError: err}
		}

		if !c.EnableAlphaFeatures {
			userResponse.Endpoints = nil
		}

		return userResponse, nil
	case http.StatusAccepted:
		if !r.AcceptsIncomplete {
			return nil, c.handleFailureResponse(response)
		}

		responseBodyObj := &bindSuccessResponseBody{}
		if err := c.unmarshalResponse(response, responseBodyObj); err != nil {
			return nil, HTTPStatusCodeError{StatusCode: response.StatusCode, ResponseError: err}
		}

		var opPtr *OperationKey
		if responseBodyObj.Operation != nil {
			opStr := *responseBodyObj.Operation
			op := OperationKey(opStr)
			opPtr = &op
		}

		userResponse := &BindResponse{
			Credentials:     responseBodyObj.Credentials,
			SyslogDrainURL:  responseBodyObj.SyslogDrainURL,
			RouteServiceURL: responseBodyObj.RouteServiceURL,
			VolumeMounts:    responseBodyObj.VolumeMounts,
			Endpoints:       responseBodyObj.Endpoints,
			Metadata:        responseBodyObj.Metadata,
			OperationKey:    opPtr,
		}
		if response.StatusCode == http.StatusAccepted {
			if c.Verbose {
				klog.Infof("broker %q: received asynchronous response", c.Name)
			}
			userResponse.Async = true
		}
		return userResponse, nil
	default:
		return nil, c.handleFailureResponse(response)
	}
}

func validateRotateBindingRequest(request *RotateBindingRequest) error {
	if request.InstanceID == "" {
		return required("instanceID")
	}

	if request.BindingID == "" {
		return required("serviceID")
	}

	if request.PredecessorBindingID == "" {
		return required("predecessorBindingID")
	}

	return nil
}
