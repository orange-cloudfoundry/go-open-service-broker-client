# `go-open-service-broker-client`

[![Build Status](https://travis-ci.org/kubernetes-sigs/go-open-service-broker-client.svg?branch=master)](https://travis-ci.org/kubernetes-sigs/go-open-service-broker-client)
[![Coverage Status](https://coveralls.io/repos/github/kubernetes-sigs/go-open-service-broker-client/badge.svg)](https://coveralls.io/github/kubernetes-sigs/go-open-service-broker-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubernetes-sigs/go-open-service-broker-client)](https://goreportcard.com/report/github.com/kubernetes-sigs/go-open-service-broker-client)
[![Godoc documentation](https://img.shields.io/badge/godoc-documentation-blue.svg)](https://godoc.org/github.com/kubernetes-sigs/go-open-service-broker-client)

A golang client for communicating with service brokers implementing the
[Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker).

## Who should use this library?

This library is most interesting if you are implementing an integration
between an application platform and the Open Service Broker API.

## Example

```go
import (
	osb "github.com/orange-cloudfoundry/go-open-service-broker-client/v2"
)

func GetBrokerCatalog(URL string) (*osb.CatalogResponse, error) {
	config := osb.DefaultClientConfiguration()
	config.URL = URL

	client, err := osb.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client.GetCatalog()
}
```

## Documentation

This client library supports the following versions of the
[Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker):

- [v2.14](https://github.com/openservicebrokerapi/servicebroker/tree/v2.14)
- [v2.13](https://github.com/openservicebrokerapi/servicebroker/tree/v2.13)
- [v2.12](https://github.com/openservicebrokerapi/servicebroker/tree/v2.12)
- [v2.11](https://github.com/openservicebrokerapi/servicebroker/tree/v2.11)

Development to support the following versions is ongoing:

- [v2.17](https://github.com/openservicebrokerapi/servicebroker/tree/v2.17)
- [v2.16](https://github.com/openservicebrokerapi/servicebroker/tree/v2.16)
- [v2.15](https://github.com/openservicebrokerapi/servicebroker/tree/v2.15)

Only fields supported by the version configured for a client are
sent/returned.

Check out the
[API specification](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md).

Check out the detailed docs for the [v2 client here](docs/).

## Current status

This repository is used in:

* The [Kubernetes `service-catalog`](https://github.com/kubernetes-sigs/service-catalog)
incubator repo
* The [`osb-broker-lib`](https://github.com/pmorie/osb-broker-lib) library for
  creating service brokers
* The [OSB Starter Pack](https://github.com/pmorie/osb-starter-pack) broker quickstart

## Goals

Overall, to make an excellent golang client for the Open Service Broker API.
Specifically:

- Provide useful insights to newcomers to the API
- Support moving between major and minor versions of the OSB API easily
- Support new auth modes in a backward-compatible manner
- Support alpha features in the Open Service Broker API in a clear manner
- Allow advanced configuration of TLS configuration to a broker
- Provide a fake client suitable for unit-type testing

Goals for the content of the project are:

- Provide high-quality godoc comments
- High degree of unit test coverage
- Code should pass vet and lint checks

## Non-goals

This project does not aim to provide:

- A v1 client
- A fake _service broker_; you may be interested in the [OSB starter
  pack](https://github.com/pmorie/osb-starter-pack)
- A conformance suite for service brokers; see
  [`osb-checker`](https://github.com/openservicebrokerapi/osb-checker) for that
- Any 'custom' API features that are not either in a released version of the
  Open Service Broker API spec or accepted into the spec but not yet released

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack channel](https://kubernetes.slack.com/messages/sig-service-catalog)
- [Mailing list](https://groups.google.com/forum/#!forum/kubernetes-sig-service-catalog)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
