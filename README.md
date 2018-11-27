# cloudflare-ddns
Add dns query to cloudflare.

DNS name is created by arguments and environ:CF_DDNS_SUBDOMAIN value.

## dependencies

- Python 3
- Internet connection(IPv4 or IPv6)
- CloudFlare accounts
- `$ pip install -r requirements.txt`

## Run

Please set environment value before running cloudflare-ddns.py

```
$ python cloudflare-ddns.py [-r [record type (default: A)]] Hostname
```
