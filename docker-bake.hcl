target "docker-metadata-action" {}

target "cloudflare-ddns" {
  inherits = ["docker-metadata-action"]
  platforms = ["linux/amd64", "linux/arm64"]
}
