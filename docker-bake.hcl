group "default" {
  targets = [
    "cloudflare-ddns",
  ]
}

target "docker-metadata-action" {}

target "cloudflare-ddns" {
  inherits = ["docker-metadata-action"]
  args = {
    target = "cloudflare-ddns"
  }
  platforms = ["linux/amd64", "linux/arm64"]
}
