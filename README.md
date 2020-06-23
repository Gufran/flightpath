# Flightpath

Flightpath has been deprecated since the release of [Ingress Gateway in Consul v1.8.0][].  
This repository is now in read only mode and will not accept community contributions.

![Test](https://github.com/Gufran/flightpath/workflows/Test/badge.svg?branch=master)

Flightpath is an Envoy Control Plane that integrates with Consul Connect and provides L7 routing
at edge.

The need for Flightpath emerged from desperately wanting to deploy Consul Connect to production
and frustration of not being able to route public traffic to a connect enabled service.  

## Getting Started

Getting started with Flightpath is unfortunately not a one command experience. Flightpath itself is not capable of handling
traffic or proxying requests. It provides a simple gRPC server for Envoy xDS API and serves configuration to Envoy. Traffic
handling and routing is all performed by Envoy.

To watch Flightpath in action you will need a basic apparatus that is easy to setup on your local machine. Please see the
[Documentation](https://docs.flightpath.xyz/) on how to setup a local test environment and go on a test flight.


## Contributing

This project no longer accepts contribution.

It remains open source under Mozilla Public License v2 and you are free to maintain your own fork.

## License

Flightpath is licensed under Mozilla Public License v2.0. You can find a copy of the license
in [LICENSE][] file or at https://www.mozilla.org/en-US/MPL/2.0/


[Envoy XDS v2 API interface]: https://www.envoyproxy.io/docs/envoy/latest/api-v2/api
[Connect Native service]: https://www.consul.io/docs/connect/native.html
[LICENSE]: https://github.com/Gufran/flightpath/blob/master/LICENSE
[Ingress Gateway in Consul v1.8.0]: https://www.consul.io/docs/connect/ingress-gateway
