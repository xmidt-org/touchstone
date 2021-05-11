# touchstone

Touchstone is an integration between go.uber.org/fx and prometheus.

[![Build Status](https://github.com/xmidt-org/touchstone/workflows/CI/badge.svg)](https://github.com/xmidt-org/touchstone/actions)
[![codecov.io](http://codecov.io/github/xmidt-org/touchstone/coverage.svg?branch=main)](http://codecov.io/github/xmidt-org/touchstone?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/xmidt-org/touchstone)](https://goreportcard.com/report/github.com/xmidt-org/touchstone)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/xmidt-org/touchstone/blob/main/LICENSE)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=xmidt-org_touchstone&metric=alert_status)](https://sonarcloud.io/dashboard?id=xmidt-org_touchstone)
[![GitHub release](https://img.shields.io/github/release/xmidt-org/touchstone.svg)](CHANGELOG.md)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/xmidt-org/touchstone)](https://pkg.go.dev/github.com/xmidt-org/touchstone)

## Summary

Touchstone provides easy bootstrapping of a prometheus client environment within a go.uber.org/fx application container.  Key features include:

- External configuration that can drive how the Registry and other components are initialized
- Simple constructors that allow individual metrics to fully participate in dependency injection
- Prebundled HTTP metrics with a simpler and more efficient instrumentation than what promhttp provides

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Details](#details)
- [Install](#install)
- [Contributing](#contributing)

## Code of Conduct

This project and everyone participating in it are governed by the [XMiDT Code Of Conduct](https://xmidt.io/code_of_conduct/). 
By participating, you agree to this Code.

## Install

go get -u github.com/xmidt-org/touchstone

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).
