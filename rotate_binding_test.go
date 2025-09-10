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
	"testing"
)

// internal message body types

const (
	testPredecessorBindingID = "test-predecessor-binding-id"
)

func defaultRotateBindingRequest() *RotateBindingRequest {
	return &RotateBindingRequest{
		InstanceID:           testInstanceID,
		BindingID:            testBindingID,
		PredecessorBindingID: testPredecessorBindingID,
	}
}

func defaultAsyncRotateBindingRequest() *RotateBindingRequest {
	r := defaultRotateBindingRequest()
	r.AcceptsIncomplete = true
	return r
}

const defaultRotateBindingRequestBody = `{"predecessor_binding_id":"test-predecessor-binding-id"}`

const successRotateBindingResponseBody = `{
  "credentials": {
	"uri": "mysql://mysqluser:pass@mysqlhost:3306/dbname",
	"username": "mysqluser",
	"password": "pass",
	"host": "mysqlhost",
	"port": 3306,
	"database": "dbname"
  }
}`

const successAsyncRotateBindingResponseBody = `{
  "operation": "test-operation-key"
}`

func successRotatebindingResponse() *BindResponse {
	return &BindResponse{
		Credentials: map[string]interface{}{
			"uri":      "mysql://mysqluser:pass@mysqlhost:3306/dbname",
			"username": "mysqluser",
			"password": "pass",
			"host":     "mysqlhost",
			"port":     float64(3306),
			"database": "dbname",
		},
	}
}

func successRotatebindingResponseAsync() *BindResponse {
	return &BindResponse{
		Async:        true,
		OperationKey: &testOperation,
	}
}

func TestRotateBinding(t *testing.T) {
	cases := []struct {
		name                string
		version             APIVersion
		enableAlpha         bool
		originatingIdentity *OriginatingIdentity
		request             *RotateBindingRequest
		httpChecks          httpChecks
		httpReaction        httpReaction
		expectedResponse    *BindResponse
		expectedErrMessage  string
		expectedErr         error
	}{
		{
			name:               "unsupported API version",
			version:            Version2_16(),
			expectedErrMessage: "RotateBinding operations are not allowed: operation not allowed: must have API version >= 2.17. Current: 2.16",
		},
		{
			name:    "invalid request",
			version: Version2_17(),
			request: func() *RotateBindingRequest {
				r := defaultRotateBindingRequest()
				r.InstanceID = ""
				return r
			}(),
			expectedErrMessage: "instanceID is required",
		},
		{
			name:    "success - created",
			version: Version2_17(),
			httpReaction: httpReaction{
				status: http.StatusCreated,
				body:   successRotateBindingResponseBody,
			},
			expectedResponse: successRotatebindingResponse(),
		},
		{
			name:    "success - asynchronous",
			version: Version2_17(),
			request: defaultAsyncRotateBindingRequest(),
			httpChecks: httpChecks{
				params: map[string]string{
					AcceptsIncomplete: "true",
				},
			},
			httpReaction: httpReaction{
				status: http.StatusAccepted,
				body:   successAsyncRotateBindingResponseBody,
			},
			expectedResponse: successRotatebindingResponseAsync(),
		},
		{
			name:    "http error",
			version: Version2_17(),
			httpReaction: httpReaction{
				err: fmt.Errorf("http error"),
			},
			expectedErrMessage: "http error",
		},
		{
			name:    "202 with no async support",
			version: Version2_17(),
			httpReaction: httpReaction{
				status: http.StatusAccepted,
				body:   successAsyncBindResponseBody,
			},
			expectedErrMessage: "Status: 202; ErrorMessage: <nil>; Description: <nil>; ResponseError: <nil>",
		},
		{
			name:    "200 with malformed response",
			version: Version2_17(),
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   malformedResponse,
			},
			expectedErrMessage: "Status: 200; ErrorMessage: <nil>; Description: <nil>; ResponseError: unexpected end of JSON input",
		},
		{
			name:    "500 with malformed response",
			version: Version2_17(),
			httpReaction: httpReaction{
				status: http.StatusInternalServerError,
				body:   malformedResponse,
			},
			expectedErrMessage: "Status: 500; ErrorMessage: <nil>; Description: <nil>; ResponseError: unexpected end of JSON input",
		},
		{
			name:    "500 with conventional failure response",
			version: Version2_17(),
			httpReaction: httpReaction{
				status: http.StatusInternalServerError,
				body:   conventionalFailureResponseBody,
			},
			expectedErr: testHTTPStatusCodeError(),
		},
	}

	for _, tc := range cases {
		if tc.request == nil {
			tc.request = defaultRotateBindingRequest()
		}

		tc.request.OriginatingIdentity = tc.originatingIdentity

		if tc.httpChecks.URL == "" {
			tc.httpChecks.URL = "/v2/service_instances/test-instance-id/service_bindings/test-binding-id"
		}

		if tc.httpChecks.body == "" {
			tc.httpChecks.body = defaultRotateBindingRequestBody
		}

		if tc.version.label == "" {
			tc.version = Version2_11()
		}

		klient := newTestClient(t, tc.name, tc.version, tc.enableAlpha, tc.httpChecks, tc.httpReaction)

		response, err := klient.RotateBinding(tc.request)

		doResponseChecks(t, tc.name, response, err, tc.expectedResponse, tc.expectedErrMessage, tc.expectedErr)
	}
}

func TestValidateRotateRequest(t *testing.T) {
	cases := []struct {
		name    string
		request *RotateBindingRequest
		valid   bool
	}{
		{
			name:    "valid",
			request: defaultRotateBindingRequest(),
			valid:   true,
		},
		{
			name: "missing binding ID",
			request: func() *RotateBindingRequest {
				r := defaultRotateBindingRequest()
				r.BindingID = ""
				return r
			}(),
			valid: false,
		},
		{
			name: "missing instance ID",
			request: func() *RotateBindingRequest {
				r := defaultRotateBindingRequest()
				r.InstanceID = ""
				return r
			}(),
			valid: false,
		},
		{
			name: "missing predecessor binding id",
			request: func() *RotateBindingRequest {
				r := defaultRotateBindingRequest()
				r.PredecessorBindingID = ""
				return r
			}(),
			valid: false,
		},
	}

	for _, tc := range cases {
		err := validateRotateBindingRequest(tc.request)
		if err != nil {
			if tc.valid {
				t.Errorf("%v: expected valid, got error: %v", tc.name, err)
			}
		} else if !tc.valid {
			t.Errorf("%v: expected invalid, got valid", tc.name)
		}
	}
}
