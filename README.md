# cloudflare-ddns

Create or Update dns record in cloudflare.

## How to use

```plain
Usage:
  cloudflare-ddns [OPTIONS] ZONE_NAME

Application Options:
  -a, --address=ADDRESS    Explicit IP address to register (overrides
                           interface/external detection)
  -d, --dry-run            Don't create or update DNS record
  -e, --external           Use external address instead of interface address
  -n, --hostname=          Name to use instead of hostname
  -i, --interface=         Interface to use address
  -4, --ipv4               Create or update only A record
  -6, --ipv6               Create or update only AAAA record
  -p, --prefix=            Prefix of hostname
  -s, --suffix=            Suffix of hostname
  -v, --verbose            Show verbose debug information
  -w, --watch=INTERVAL     Run continuously, updating every INTERVAL (e.g. 5m,
                           1h)

Help Options:
  -h, --help               Show this help message
```

The record name will be of the form `[prefix.]hostname.[suffix.]ZONE_NAME`.

### Watch mode

Use `--watch` to run continuously and update DNS records at a fixed interval:

```sh
CLOUDFLARE_API_TOKEN=xxx cloudflare-ddns --watch=5m example.com
```

The program syncs immediately on startup, then repeats every interval. It exits cleanly on SIGTERM or SIGINT.

### Kubernetes DaemonSet

A DaemonSet manifest is provided at `deploy/kubernetes/daemonset.yaml`. It uses the Kubernetes Downward API to register each node's name and IP automatically:

- `spec.nodeName` → `--hostname`
- `status.hostIP` → `--address`

```sh
# Edit REPLACE_ME (API token) and example.com (zone name) before applying
kubectl apply -f deploy/kubernetes/daemonset.yaml
```

## dev dependencies

- Go

## dependencies

- Internet connection(IPv4 or IPv6)
- Cloudflare accounts

## How to build

- `go mod download`
- `go build`

## License

Copyright (c) 2022 Rokoucha

Released under the MIT license, see LICENSE.
