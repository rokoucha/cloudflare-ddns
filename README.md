# cloudflare-ddns

Create or Update dns record in cloudflare.

## How to use

```plain
Usage:
  cloudflare-ddns [OPTIONS] ZONE_NAME

Application Options:
  -d, --dry-run    Don't create or update DNS record
  -e, --external   Use external address instead of interface address
  -n, --hostname=  Name to use instead of hostname
  -i, --interface= Interface to use address
  -4, --ipv4       Create or update only A record
  -6, --ipv6       Create or update only AAAA record
  -p, --prefix=    Prefix of hostname
  -s, --suffix=    Suffix of hostname

Help Options:
  -h, --help       Show this help message
```

The record name will be of the form `[prefix.]hostname.[suffix.]ZONE_NAME`.

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
