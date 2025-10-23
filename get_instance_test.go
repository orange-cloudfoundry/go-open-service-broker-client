/*
Copyright 2019 The Kubernetes Authors.

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

const okInstanceBytes = `{
  "service_id": "test-service",
  "plan_id": "test-plan"
}`

func defaultGetInstanceRequest() *GetInstanceRequest {
	return &GetInstanceRequest{
		InstanceID: testInstanceID,
		ServiceID:  testServiceID,
		PlanID:     testPlanID,
	}
}

func okGetInstanceResponse() *GetInstanceResponse {
	return &GetInstanceResponse{
		ServiceID: "test-service",
		PlanID:    "test-plan",
	}
}

func getInstanceResponseWithMetadata() *GetInstanceResponse {
	return &GetInstanceResponse{
		ServiceID: "test-service",
		PlanID:    "test-plan",
		Metadata:  ServiceInstanceMetadata{},
	}
}

func TestGetInstance(t *testing.T) {
	cases := []struct {
		name               string
		enableAlpha        bool
		request            *GetInstanceRequest
		APIVersion         APIVersion
		httpReaction       httpReaction
		expectedResponse   *GetInstanceResponse
		expectedErrMessage string
		expectedErr        error
	}{
		{
			name: "success",
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   okInstanceBytes,
			},
			expectedResponse: okGetInstanceResponse(),
		},
		{
			name: "http error",
			httpReaction: httpReaction{
				err: fmt.Errorf("http error"),
			},
			expectedErrMessage: "http error",
		},
		{
			name: "200 with malformed response",
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   malformedResponse,
			},
			expectedErrMessage: "Status: 200; ErrorMessage: <nil>; Description: <nil>; ResponseError: unexpected end of JSON input",
		},
		{
			name: "500 with malformed response",
			httpReaction: httpReaction{
				status: http.StatusInternalServerError,
				body:   malformedResponse,
			},
			expectedErrMessage: "Status: 500; ErrorMessage: <nil>; Description: <nil>; ResponseError: unexpected end of JSON input",
		},
		{
			name: "500 with conventional response",
			httpReaction: httpReaction{
				status: http.StatusInternalServerError,
				body:   conventionalFailureResponseBody,
			},
			expectedErr: testHTTPStatusCodeError(),
		},
		{
			name:               "unsupported API version",
			APIVersion:         Version2_13(),
			expectedErrMessage: "GetInstance not allowed: operation not allowed: must have API version >= 2.14. Current: 2.13",
		},
	}

	for _, tc := range cases {
		if tc.request == nil {
			tc.request = defaultGetInstanceRequest()
		}

		httpChecks := httpChecks{
			URL: "/v2/service_instances/test-instance-id",
		}

		if tc.APIVersion.label == "" {
			tc.APIVersion = LatestAPIVersion()
		}

		klient := newTestClient(t, tc.name, tc.APIVersion, tc.enableAlpha, httpChecks, tc.httpReaction)

		response, err := klient.GetInstance(tc.request)

		doResponseChecks(t, tc.name, response, err, tc.expectedResponse, tc.expectedErrMessage, tc.expectedErr)
	}
}
