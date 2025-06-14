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
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"k8s.io/klog/v2"
)

const (
	// APIVersionHeader is the header value associated with the version of the Open
	// Service Broker API version.
	APIVersionHeader = "X-Broker-API-Version"
	// OriginatingIdentityHeader is the header associated with originating
	// identity.
	OriginatingIdentityHeader = "X-Broker-API-Originating-Identity"
	// RequestIdentityFeader is the header associated with request identity
	RequestIdentityheader = "X-Broker-API-Request-Identity"
	// PollingDelayHeader is the header used by the brokers to tell the clients
	// how many seconds they should wait before retrying the polling
	PollingDelayHeader = "Retry-After"

	catalogURL                 = "%s/v2/catalog"
	serviceInstanceURLFmt      = "%s/v2/service_instances/%s"
	lastOperationURLFmt        = "%s/v2/service_instances/%s/last_operation"
	bindingLastOperationURLFmt = "%s/v2/service_instances/%s/service_bindings/%s/last_operation"
	bindingURLFmt              = "%s/v2/service_instances/%s/service_bindings/%s"
)

// NewClient is a CreateFunc for creating a new functional Client and
// implements the CreateFunc interface.
func NewClient(config *ClientConfiguration) (Client, error) {
	httpClient := &http.Client{
		Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
	}

	// use default values lifted from DefaultTransport
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if config.TLSConfig != nil {
		transport.TLSClientConfig = config.TLSConfig
	} else {
		transport.TLSClientConfig = &tls.Config{}
	}
	if config.Insecure {
		transport.TLSClientConfig.InsecureSkipVerify = true
	}
	if len(config.CAData) != 0 {
		if transport.TLSClientConfig.RootCAs == nil {
			transport.TLSClientConfig.RootCAs = x509.NewCertPool()
		}
		transport.TLSClientConfig.RootCAs.AppendCertsFromPEM(config.CAData)
	}
	if transport.TLSClientConfig.InsecureSkipVerify && transport.TLSClientConfig.RootCAs != nil {
		return nil, errors.New("Cannot specify root CAs and to skip TLS verification")
	}
	httpClient.Transport = transport

	c := &client{
		Name:                config.Name,
		URL:                 strings.TrimRight(config.URL, "/"),
		APIVersion:          config.APIVersion,
		EnableAlphaFeatures: config.EnableAlphaFeatures,
		Verbose:             config.Verbose,
		httpClient:          httpClient,
	}
	c.doRequestFunc = c.doRequest

	if config.AuthConfig != nil {
		if config.AuthConfig.BasicAuthConfig == nil && config.AuthConfig.BearerConfig == nil {
			return nil, errors.New("Non-nil AuthConfig cannot be empty")
		}
		if config.AuthConfig.BasicAuthConfig != nil && config.AuthConfig.BearerConfig != nil {
			return nil, errors.New("Only one AuthConfig implementation must be set at a time")
		}

		c.AuthConfig = config.AuthConfig
	}

	return c, nil
}

var _ CreateFunc = NewClient

type doRequestFunc func(request *http.Request) (*http.Response, error)

// client provides a functional implementation of the Client interface.
type client struct {
	Name                string
	URL                 string
	APIVersion          APIVersion
	AuthConfig          *AuthConfig
	EnableAlphaFeatures bool
	Verbose             bool

	httpClient    *http.Client
	doRequestFunc doRequestFunc
}

var _ Client = &client{}

// This file contains shared methods used by each interface method of the
// Client interface.  Individual interface methods are in the following files:
//
// GetCatalog: get_catalog.go
// ProvisionInstance: provision_instance.go
// UpdateInstance: update_instance.go
// DeprovisionInstance: deprovision_instance.go
// PollLastOperation: poll_last_operation.go
// Bind: bind.go
// Unbind: unbind.go
// RotateBinding: rotate_binding.go

const (
	contentType = "Content-Type"
	jsonType    = "application/json"
)

// prepareAndDo prepares a request for the given method, URL, and
// message body, and executes the request, returning an http.Response or an
// error.  Errors returned from this function represent http-layer errors and
// not errors in the Open Service Broker API.
func (c *client) prepareAndDo(method, URL string, params map[string]string, body interface{}, originatingIdentity *OriginatingIdentity) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		bodyReader = bytes.NewReader(bodyBytes)
	}

	request, err := http.NewRequest(method, URL, bodyReader)
	if err != nil {
		return nil, err
	}

	request.Header.Set(APIVersionHeader, c.APIVersion.HeaderValue())
	if bodyReader != nil {
		request.Header.Set(contentType, jsonType)
	}

	if c.AuthConfig != nil {
		if c.AuthConfig.BasicAuthConfig != nil {
			basicAuth := c.AuthConfig.BasicAuthConfig
			request.SetBasicAuth(basicAuth.Username, basicAuth.Password)
		} else if c.AuthConfig.BearerConfig != nil {
			bearer := c.AuthConfig.BearerConfig
			request.Header.Set("Authorization", "Bearer "+bearer.Token)
		}
	}

	requestId := uuid.New()
	request.Header.Set(RequestIdentityheader, requestId.String())

	if c.APIVersion.AtLeast(Version2_13()) && originatingIdentity != nil {
		headerValue, err := buildOriginatingIdentityHeaderValue(originatingIdentity)
		if err != nil {
			return nil, err
		}
		request.Header.Set(OriginatingIdentityHeader, headerValue)
	}

	if params != nil {
		q := request.URL.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		request.URL.RawQuery = q.Encode()
	}

	if c.Verbose {
		klog.Infof("broker %q: doing request to %q", c.Name, URL)
	}

	return c.doRequestFunc(request)
}

func (c *client) doRequest(request *http.Request) (*http.Response, error) {
	return c.httpClient.Do(request)
}

// unmarshalResponse unmarshals the response body of the given response into
// the given object or returns an error.
func (c *client) unmarshalResponse(response *http.Response, obj interface{}) error {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if c.Verbose {
		klog.Infof("broker %q: response body: %v, type: %T", c.Name, string(body), obj)
	}

	err = json.Unmarshal(body, obj)
	if err != nil {
		return err
	}

	return nil
}

// handleFailureResponse returns an HTTPStatusCodeError for the given
// response.
func (c *client) handleFailureResponse(response *http.Response) error {
	klog.Info("handling failure responses")

	httpErr := HTTPStatusCodeError{
		StatusCode: response.StatusCode,
	}

	brokerResponse := make(map[string]interface{})
	if err := c.unmarshalResponse(response, &brokerResponse); err != nil {
		httpErr.ResponseError = err
		return httpErr
	}

	if errorMessage, ok := brokerResponse["error"].(string); ok {
		httpErr.ErrorMessage = &errorMessage
	}

	if description, ok := brokerResponse["description"].(string); ok {
		httpErr.Description = &description
	}

	return httpErr
}

func buildOriginatingIdentityHeaderValue(i *OriginatingIdentity) (string, error) {
	if i == nil {
		return "", nil
	}
	if i.Platform == "" {
		return "", errors.New("originating identity platform must not be empty")
	}
	if i.Value == "" {
		return "", errors.New("originating identity value must not be empty")
	}
	if err := isValidJSON(i.Value); err != nil {
		return "", fmt.Errorf("originating identity value must be valid JSON: %v", err)
	}
	encodedValue := base64.StdEncoding.EncodeToString([]byte(i.Value))
	headerValue := fmt.Sprintf("%v %v", i.Platform, encodedValue)
	return headerValue, nil
}

func isValidJSON(s string) error {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js)
}

// validateClientVersionIsAtLeast returns an error if client version is not at
// least the specified version
func (c *client) validateClientVersionIsAtLeast(version APIVersion) error {
	if !c.APIVersion.AtLeast(version) {
		return OperationNotAllowedError{
			reason: fmt.Sprintf(
				"must have API version >= %s. Current: %s",
				version,
				c.APIVersion.label,
			),
		}
	}

	return nil
}

// drainReader reads and discards the remaining data in reader (for example
// response body data) For HTTP this ensures that the http connection
// could be reused for another request if the keepalive is enabled.
// see https://gist.github.com/mholt/eba0f2cc96658be0f717#gistcomment-2605879
// Not certain this is really needed here for the Broker vs a http server
// but seems safe and worth including at this point
func drainReader(reader io.Reader) error {
	if reader == nil {
		return nil
	}
	_, drainError := io.Copy(io.Discard, io.LimitReader(reader, 4096))
	return drainError
}

// internal message body types

type asyncSuccessResponseBody struct {
	Operation *string `json:"operation"`
}
