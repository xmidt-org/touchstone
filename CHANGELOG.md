# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
- touchbundle package supports creation of metric bundles, which are logical groups of metrics
- a version of NewUntypedFunc that allows for flexible function signatures
- WrapUntypedFunc exposes the basic logic around wrapping closures for metrics
- touchbundle now supports untyped metrics

## [v0.1.0]
- Updated go.uber.org/fx to 1.17.1
- Updated github.com/prometheus/client_golang to 1.12.1
- Remove use of fx.Printer for messages.  Replaced with an optional zap.Logger.
- Fx components are now grouped under a common module.

### Fixed
- Broken README badge links. [#12](https://github.com/xmidt-org/touchstone/pull/12)

## [v0.0.3]

### Added
- utility methods for dealing with prometheus.AlreadyRegisteredError
- dynamically-typed metric Factory methods

## [v0.0.2]

### Added
- build info collector
- go-kit integration [#6](https://github.com/xmidt-org/touchstone/pull/6)
- touchtest package with useful testing assertions and utilities

## [v0.0.1]
- Initial creation
- external configuration
- bootstrapping for the core prometheus objects:  Registerer and Gatherer
- bootstrapping for the HTTP environment
- bundled HTTP metrics

[Unreleased]: https://github.com/xmidt-org/touchstone/compare/v0.1.0..HEAD
[v0.1.0]: https://github.com/xmidt-org/touchstone/compare/v0.0.3...v0.1.0
[v0.0.3]: https://github.com/xmidt-org/touchstone/compare/v0.0.2...v0.0.3
[v0.0.2]: https://github.com/xmidt-org/touchstone/compare/v0.0.1...v0.0.2
[v0.0.1]: https://github.com/xmidt-org/touchstone/releases/tag/v0.0.1
