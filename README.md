# cloudflare-ddns

Create or Update dns record in cloudflare.

## How to use

```
Usage:
  cloudflare-ddns [OPTIONS]

Application Options:
  -d, --dry-run    dry run
  -e, --external   Use external global address
  -n, --hostname=  hostname
  -i, --interface= interface name
  -4, --ipv4       Add A record
  -6, --ipv6       Add AAAA record
  -p, --prefix=    Prefix
  -s, --subdomain= Subdomain

Help Options:
  -h, --help       Show this help message
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
