group "default" {
  targets = [
    "cloudflare-ddns",
  ]
}

target "docker-metadata-action" {}

target "cloudflare-ddns" {
  inherits = ["docker-metadata-action"]
  tags     = make_tags("cloudflare-ddns")
  args = {
    target = "cloudflare-ddns"
  }
  platforms = ["linux/amd64", "linux/arm64"]
}

variable "DOCKER_METADATA_OUTPUT_TAGS" {
  default = ""
}
function "make_tags" {
  params = [ns]
  result = split("\n", replace("${DOCKER_METADATA_OUTPUT_TAGS}", ":", "/${ns}:"))
}
