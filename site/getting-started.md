# Getting Started

There are several ways to install Flightpath.

 - Docker Image
 - Executable Binary
 - Building From Source

## Installation

### Docker Image

Docker images are available for each release, you can pull the image for any tag using 

```shell
# pull the image for release tag 1.2.3
docker pull gufran/flightpath:1.2.3
```

or you can also pull an image built from the head of master branch using `latest` tag.
It is not recommended to use `latest` docker image in production since the master branch is not guaranteed to be stable
or even in a working state.

!!! caution
    Always use a specific release tag in production. The `latest` docker image always contains the most
    recent commits to master branch and expected to be highly unstable.

### Executable Binary

Executable binaries can be downloaded from [Github Releases][] page.
Note that only a tagged release is available for download from the release page. There are unstable binaries that are
built from master branch and can be downloaded from [Github Actions Artifacts][] page. Simply select the build you want
and download the binaries.

!!! info
    Executable binaries are only available for Linux, MacOS and Windows operating systems with
    AMD64 architecture. If you want to run Flightpath on other OS or CPU architecture you should
    consider building from source.

### Building From Source

To build from source you need to have Golang v1.13 or better installed on your machine.
You can build Flightpath using

```shell
go get github.com/Gufran/flightpath
```

Binary built with `go get` will not have the build metadata embedded in it. If you care for build metadata then you
should build with following command:

```shell
git clone https://github.com/Gufran/flightpath.git
cd flightpath
bash build.sh #-> will build ./flightpath
```

## Test Flight

Flightpath depends on Consul Catalog to generate Envoy configuration. You are going to need Consul version 1.6 or better
and Envoy version 1.12 or better installed before you can run Flightpath.

!!! caution
    A docker based setup does not work properly on OSX because Docker on OSX runs on a virtualized linux host.
    This added layer of virtualization causes network problems in both bridged and host mode. It is possible to
    make it work on OSX with some effort but for now Docker is not involved.

The quick start script uses [http-echo][] to run a simple web server. You can pull it with `go get`:

```shell
go get github.com/hashicorp/http-echo
```

On OSX you can install Consul and Envoy using Homebrew

```shell
brew install consul envoy
```

Once you've ensured the dependencies you can use [internal/testflight.sh][] script to spin up an environment with following
components:

 - [x] Consul agent
 - [x] HTTP service without connect sidecar
 - [x] HTTP service with connect sidecar
 - [x] Flightpath
 - [x] Envoy

At this point you should have the address of consul UI and both the services printed in your terminal.
If you go to the [Consul UI][] you should see the web services and flightpath registered in the catalog. 

You can now navigate to [Service Without Connect][] or [Service With Connect][] and be greeted with the static text
response from http-echo service.

Feel free to browse around in [Consul UI][] or read the [internal/testflight.sh] script to get an understanding of what
is going on.












[Github Releases]: https://github.com/Gufran/flightpath/releases
[Github Actions Artifacts]: https://github.com/Gufran/flightpath/actions?query=workflow%3ATest
[http-echo]: https://github.com/hashicorp/http-echo
[internal/testflight.sh]: https://github.com/Gufran/flightpath/blob/master/internal/testflight.sh
[Consul UI]: http://localhost:8500/ui/dc1/services
[Service Without Connect]: http://without-connect.app.local:9292/
[Service With Connect]: http://with-connect.app.local:9292/