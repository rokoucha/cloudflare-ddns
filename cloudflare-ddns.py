from CloudFlare import CloudFlare
from os import environ, uname
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
                print('UNCHANGED: %s %s %s' % (dns_name, target_type, ip_address))
                return

            # {hostname}.CF_DDNS_SUBDOMAIN.CF_ZONE_NAME(A and AAAA)
            cf.zones.dns_records.put(zone_id, record_id, data={
                'name': dns_name,
                'type': target_type,
                'content': ip_address
            })

            print('UPDATED: %s %s %s -> %s' % (dns_name, target_type, record_ip, ip_address))
    else:
        # Create
        cf.zones.dns_records.post(zone_id, data={
            'name': dns_name,
            'type': target_type,
            'content': ip_address
        })

        print('CREATED: %s %s %s' % (dns_name, target_type, ip_address))


def main():
    zone_name = environ["CF_ZONE_NAME"] 
    dns_name = '{}.{}'.format(uname()[1], environ["CF_DDNS_SUBDOMAIN"] )

    # Get IP Address
    target_types = []
    try:
        ipv4_address = get('https://v4.ident.me').text
        target_types.append('A')
    except exceptions.RequestException:
        ipv4_address = ''
    try:
        ipv6_address = get('https://v6.ident.me').text
        target_types.append('AAAA')
    except exceptions.RequestException:
        ipv6_address = ''

    # Update record
    for target_type in target_types:
        ip_address = ipv4_address if target_type is 'A' else ipv6_address
        ip_version = 'v4' if target_type is 'A' else 'v6'

        # A and AAAA record
        post_record(zone_name, dns_name, target_type, ip_address)

        # A or AAAA only record
        post_record(zone_name, '{}.{}'.format(ip_version,dns_name), target_type, ip_address)

if __name__ == '__main__':
    main()
