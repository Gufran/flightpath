job "without-connect" {
  datacenters = ["dc1"]
  group "api" {
    network {
      mode = "bridge"
      port "api" {}
    }

    service {
      name = "without-connect"
      port = "api"
      tags = ["in-flightpath"]

      meta {
        flightpath-route-main = "without-connect.app.local"
      }
    }

    task "without-connect" {
      driver = "docker"
      config {
        image   = "hashicorp/http-echo:latest"
        args    = ["-listen", ":${NOMAD_PORT_api}", "-text", "Echo from the service without connect"]
      }
    }
  }
}
