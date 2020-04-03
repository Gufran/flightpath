#!/bin/bash

set -eou pipefail

function die() {
  echo "${@}"
  exit 1
}

[ ! -d ./internal ] && die "Please execute the script from project root"

flightpath_exec='flightpath'

if ! command -v "${flightpath_exec}" >/dev/null 2>&1; then
  if [ -x flightpath ]; then
    flightpath_exec='./flightpath'
  else
    read -r -n 1 -p "flightpath executable is not available. Do you want to build from source? [y/N] " build_from_source
    [ "${build_from_source}" != 'y' ] && die "flightpath executable is not available"
    /bin/bash build.sh native
  fi
fi

if ! command -v consul >/dev/null 2>&1; then
  die "Consul is not available"
fi

if ! command -v envoy >/dev/null 2>&1; then
  die "Envoy is not available"
fi

tmpdir=$(mktemp -d -p "${PWD}")
consul_logs_file="${tmpdir}/consul.log"
flightpath_logs_file="${tmpdir}/flightpath.log"
envoy_logs_file="${tmpdir}/envoy.log"
envoy_access_logs="${tmpdir}/access.log"
processes=""

function cleanup() {
  trap '' INT TERM

  echo "Terminating PIDs: ${processes}"
  for each in ${processes}; do
    kill -TERM $each
  done

  rm -rf "${tmpdir}"
}
trap cleanup EXIT

echo "Starting consul agent. Logs will be streamed to ${consul_logs_file#${PWD}/}"
consul agent -dev -client 0.0.0.0 -config-dir ./internal/consul-config >"${consul_logs_file}" 2>&1 &
consul_pid="${!}"
echo "Consul started with process ID ${consul_pid}"
processes="${processes} ${consul_pid}"

until consul members >/dev/null 2>&1; do
  echo "Waiting for consul to boot..."
  sleep 3
done

echo "Starting local services"
http-echo -listen :8181 -text "Hello from the service without a sidecar" >"${tmpdir}/service-without-connect.log" 2>&1 &
s1pid="${!}"
echo "Service without connect started with process ID ${s1pid}"
processes="${processes} ${s1pid}"

until curl --fail http://127.0.0.1:8181/health > /dev/null 2>&1; do
  echo "Waiting for service to come online..."
  sleep 3
done

http-echo -listen :8182 -text "Hello from the service with connect sidecar" >"${tmpdir}/service-with-connect.log" 2>&1 &
s2pid="${!}"
echo "Service with connect sidecar started with process ID ${s2pid}"
processes="${processes} ${s2pid}"

until curl --fail http://127.0.0.1:8182/health > /dev/null 2>&1; do
  echo "Waiting for service to come online..."
  sleep 3
done

consul connect envoy -sidecar-for with-connect >"${tmpdir}/sidecar.log" 2>&1 &
sidecar_pid="${!}"
echo "Sidecar started with process ID ${sidecar_pid}"
processes="${processes} ${sidecar_pid}"

echo "Starting Flightpath. Logs will be streamed to ${flightpath_logs_file#${PWD}/}"
"${flightpath_exec}" -metrics.sink=stderr -envoy.http.access-logs "${envoy_access_logs}" >"${flightpath_logs_file}" 2>&1 &
flightpath_pid="${!}"
echo "Flightpath started with process ID ${flightpath_pid}"
processes="${processes} ${flightpath_pid}"

echo "Starting Envoy"
echo "Envoy logs will be streamed to ${envoy_logs_file#${PWD}/}"
echo "Access logs will be streamed to ${envoy_access_logs#${PWD}/}"
envoy -c internal/envoy-config.yaml --log-format "%v" > "${envoy_logs_file}" 2>&1 &
envoy_pid="${!}"
echo "Envoy started with process ID ${envoy_pid}"
processes="${processes} ${envoy_pid}"

echo "==================================================================="
echo "             All components have been started"
echo
echo "               Consul UI: http://127.0.0.1:8500"
echo " Service Without Connect: http://without-connect.app.local:9292/"
echo "    Service With Connect: http://with-connect.app.local:9292/"
echo "          Logs directory: ${tmpdir#${PWD}/}"
echo "                All PIDs: ${processes}"
echo
echo "             Press ctrl+c to quit everything"
echo "==================================================================="

cat
