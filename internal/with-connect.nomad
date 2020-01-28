job "with-connect" {
  datacenters = ["dc1"]
  group "api" {
    network {
      mode = "bridge"
      port "api" {}
    }

    service {
      name = "with-connect"
      port = "api"
      tags = ["in-flightpath"]

      connect {
        sidecar_service {}
      }

      meta {
        flightpath-route-main = "with-connect.app.local"
      }
    }

    task "with-connect" {
      driver = "docker"
      config {
        image = "hashicorp/http-echo:latest"
        args    = ["-listen", ":${NOMAD_PORT_api}", "-text", "Echo from the service with connect sidecar"]
      }
    }
  }
}
