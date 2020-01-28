job "without-connect" {
  datacenters = ["dc1"]

  update {
    max_parallel      = 5
    health_check      = "checks"
    min_healthy_time  = "5s"
    healthy_deadline  = "1m"
    progress_deadline = "3m"
    auto_revert       = true
    auto_promote      = true
    canary            = 5
  }

  meta {
    version = "8"
  }

  group "api" {
    count = 5

    task "without-connect" {
      driver = "docker"
      config {
        image = "hashicorp/http-echo:latest"
        args    = ["-listen", ":${NOMAD_PORT_api}", "-text", "Service without sidecar version ${NOMAD_META_version} alloc ${NOMAD_ALLOC_ID}"]
      }

      resources {
        network {
          port "api" {}
        }
      }

      service {
        name = "without-connect"
        port = "api"
        tags = ["in-flightpath"]
        meta {
          flightpath-route-main = "without-connect.app.local"
        }
      }
    }
  }
}
