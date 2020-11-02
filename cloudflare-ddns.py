from CloudFlare import CloudFlare
from argparse import ArgumentParser
from os import environ
from requests import exceptions, get

def post_record(zone_name, dns_name, target_type, ip_address):
    cf = CloudFlare()

    # Get zone
    zone_info = cf.zones.get(params={'name': zone_name})
    if len(zone_info) > 0:
        if 'id' in zone_info[0]:
            zone_id = zone_info[0]['id']
        else:
            raise ValueError('Zone is not found')

    # Get record
    dns_records = cf.zones.dns_records.get(zone_id, params={'name': dns_name + '.' + zone_name, 'type': target_type})

    # Is record exist?
    if len(dns_records) > 0:
        # Update
        dns_record = dns_records[0]
        if 'id' in dns_record:
            record_id = dns_record['id']
            record_ip = dns_record['content']

            if ip_address==record_ip:
                print('UNCHANGED: {} -> {} {} {}'.format(zone_name, dns_name, target_type, ip_address))
                return

            # {hostname}.CF_DDNS_SUBDOMAIN.CF_ZONE_NAME(A and AAAA)
            cf.zones.dns_records.put(zone_id, record_id, data={
                'name': dns_name,
                'type': target_type,
                'content': ip_address
            })

            print('CHANGED: {} -> {} {} {} -> {}'.format(zone_name, dns_name, target_type, record_ip, ip_address))
    else:
        # Create
        cf.zones.dns_records.post(zone_id, data={
            'name': dns_name,
            'type': target_type,
            'content': ip_address
        })

        print('CREATED: {} -> {} {} {}'.format(zone_name, dns_name, target_type, ip_address))


def main():
    parser = ArgumentParser()

    parser.add_argument('-r', '--record', action='store', default='A', help='Record type')
    parser.add_argument('name', help='DNS Name')
    args = parser.parse_args()

    zone_name = environ["CF_ZONE_NAME"]
    dns_name = '{}.{}'.format(args.name, environ["CF_DDNS_SUBDOMAIN"] )

    # Get IP Address
    try:
        ipv4_address = get('https://v4.ident.me').text
    except exceptions.RequestException:
        ipv4_address = ''
    try:
        ipv6_address = get('https://v6.ident.me').text
    except exceptions.RequestException:
        ipv6_address = ''

    ip_address = ipv4_address if args.record == 'A' else ipv6_address

    # Update record
    post_record(zone_name, dns_name, args.record, ip_address)

if __name__ == '__main__':
    main()
