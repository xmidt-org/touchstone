# touchstone

Touchstone is an integration between [go.uber.org/fx](https://pkg.go.dev/go.uber.org/fx) and [prometheus](https://prometheus.io/).

[![Build Status](https://github.com/xmidt-org/touchstone/actions/workflows/ci.yml/badge.svg)](https://github.com/xmidt-org/touchstone/actions/workflows/ci.yml)
[![codecov.io](http://codecov.io/github/xmidt-org/touchstone/coverage.svg?branch=main)](http://codecov.io/github/xmidt-org/touchstone?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/xmidt-org/touchstone)](https://goreportcard.com/report/github.com/xmidt-org/touchstone)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=xmidt-org_touchstone&metric=alert_status)](https://sonarcloud.io/dashboard?id=xmidt-org_touchstone)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/xmidt-org/touchstone/blob/main/LICENSE)
[![GitHub Release](https://img.shields.io/github/release/xmidt-org/touchstone.svg)](https://github.com/xmidt-org/touchstone/releases)
[![GoDoc](https://pkg.go.dev/badge/github.com/xmidt-org/touchstone)](https://pkg.go.dev/github.com/xmidt-org/touchstone)

## Summary

Touchstone provides easy bootstrapping of a prometheus client environment within a
go.uber.org/fx application container.  Key features include:

- External configuration that drives how the `prometheus.Registerer` and `prometheus.Gatherer` are initialized
- Simple constructors (`Counter`, `GaugeVec`, `Histogram`, etc.) that allow individual metrics to fully participate in dependency injection
- A `Factory` for creating metrics dynamically, with helpers for handling `prometheus.AlreadyRegisteredError`
- Optional `zap.Logger` integration for diagnostic output

### Packages

| Package | Purpose |
|---------|---------|
| `touchstone` | Core fx bootstrapping for the prometheus `Registerer` and `Gatherer`, plus metric constructors |
| `touchhttp` | HTTP metrics handler configuration and prebundled server and client instrumentation that is simpler and more efficient than `promhttp` |
| `touchbundle` | Declarative bundles of metrics driven by struct tags, usable for both creating application metrics and verifying them in tests |
| `touchkit` | Adapters that let code written against [go-kit metrics](https://pkg.go.dev/github.com/go-kit/kit/metrics) participate in touchstone dependency injection |
| `touchtest` | Testing assertions and utilities for prometheus metrics |

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Install](#install)
- [Examples](#examples)
- [Releases](#releases)
- [Contributing](#contributing)

## Code of Conduct

This project and everyone participating in it are governed by the [XMiDT Code Of Conduct](https://xmidt.io/code_of_conduct/).
By participating, you agree to this Code.

## Install

```
go get github.com/xmidt-org/touchstone
```

## Examples

Runnable examples for each package are available in the [GoDoc](https://pkg.go.dev/github.com/xmidt-org/touchstone).

## Releases

Releases are produced automatically by [goreleaser](https://goreleaser.com/).  Release notes are generated
from the commit history and published on the [GitHub Releases](https://github.com/xmidt-org/touchstone/releases) page.

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).
