# Changelog

## v0.0.3

This is a catch up release to fix the CI builds that were broken by enabling
vendoring mode on build and tests. It does not contain any change in code
apart from removing `-mod=vendor` from build script.

Use this release over `0.0.2` as the previous automated build did not
publish all artifacts.

## v0.0.2

### Added

 - Flightpath exposes go runtime telemetry to metrics sink if enabled
 - Flightpath can configure Envoy to initiate request tracing if enabled
 - Added example systemd unit files in `contrib/systemd`
 - Improved sample Envoy configuration file
   - Demonstrate metrics and tracing configuration
   - Static health check endpoint that does not rely on admin inteface
   - Configuration to tag metrics

### Changed

 - Debug server no longer runs by default

## v0.0.1

### Added

 - Update `internal/envoy-config.yaml` to demonstrate telemetry configuration

### Fixed

 - Immediately remove nodes with failing consul health checks
 - Only send updates to the xDS server when there is an actual change

### Changed

 - Support printing metrics to stderr.
   - `-dogstatsd` argument is no longer available
   - A new argument `-metrics.sink` can be used to configure where to send the metrics


