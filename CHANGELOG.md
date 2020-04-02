# Changelog

## v0.0.5 (Unreleased)

### Breaking Changes

 - `-envoy.access-logs` has been changed to `-envoy.http.access-logs`

### Added

 - Command line options to change the Envoy listener configuration
   - See all `-envoy.listen.*` and `-envoy.http.*` options for Listener and HTTP manager configuration that can be changed
 - Upstream cluster and route configuration can be controlled using service metadata in consul catalog
 - A vagrant configuration is now available to run flightpath e2e tests

## v0.0.4

### Fixed

 - Run loop to emit runtime metrics in a separate goroutine
 - Only enable runtime metrics when the argument is set

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


