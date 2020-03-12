# Changelog

## Unreleased

### Added

 - Update `internal/envoy-config.yaml` to demonstrate telemetry configuration

### Fixed

 - Immediately remove nodes with failing consul health checks
 - Only send updates to the xDS server when there is an actual change

### Changed

 - Support printing metrics to stderr.
   - `-dogstatsd` argument is no longer available
   - A new argument `-metrics.sink` can be used to configure where to send the metrics


