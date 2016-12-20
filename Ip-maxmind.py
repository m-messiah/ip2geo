import base64
import codecs
from collections import namedtuple
import csv
import os
import io
import sys
import zipfile
import requests

import iptools

IP_VERSION = {'ipv4': 'IPv4', 'ipv6': 'IPv6'}
# Include or Exclude
ei_action = None


def error(text):
    sys.stderr.write(text + "\n")
    exit()

archive = requests.get('http://geolite.maxmind.com/download/geoip/database/GeoLite2-City-CSV.zip')

with zipfile.ZipFile(io.BytesIO(archive.content)) as ziped_data:

    # get parent folder name
    zip_name = ziped_data.namelist()[0].split('/')[0]

    # get list of available csv languages
    available_languages = [x.replace('GeoLite2-City-Locations-', '').replace('.csv', '').replace(zip_name+'/', '')
                           for x in ziped_data.namelist() if 'GeoLite2-City-Locations' in x]

    # arg testing
    # args: lang ip_ver path include-exclude ISO_CODEs
    args = sys.argv[1:]
    if len(args) < 3:
        error("\nUsage:python3 Ip-maxmind.py language ip_ver path -optional(include or exclude one or more country)\n"
              "\nExample:python3 Ip-maxmind.py ru ipv4 /full/path/to/file.txt"
              "\n\nExample:python3 Ip-maxmind.py ru ipv4 /full/path/to/file.txt include RU UA BY"
              "\navailable_languages:" + ' '.join(available_languages))
    if args[0] not in available_languages:
        error('Unsuported language:' + args[0])

    filename = args[2]

    if len(args) > 3:
        if args[3] != 'include' and args[3] != 'exclude':
            error('4rd argument must be a include or exclude, not ' + args[3])
        if len(args) <= 4:
            error('nothing to ' + args[3])
        else:
            ei_action = args[3]
            iso_codes = args[4:]  # for include or exclude

    if IP_VERSION.get(args[1]):
        ipver = IP_VERSION.get(args[1])
    else:
        error('ip_version can be ipv4 or ipv6 not ' + args[1])

    with ziped_data.open(os.path.join(zip_name, 'GeoLite2-City-Blocks-' + ipver + '.csv')) as zip_blocks:
        city_blocks = [i for i in csv.reader(codecs.iterdecode(zip_blocks, 'utf-8'), delimiter=',')]
        # field names from 1st row of csv file
        Block_IP = namedtuple('Block_'+ipver, ' '.join(city_blocks[0]))
        ip_blocks = [Block_IP(*x) for x in city_blocks[1:]]

    with ziped_data.open(os.path.join(zip_name, 'GeoLite2-City-Locations-'+args[0]+'.csv')) as zip_city_locations:
        city_locations = [i for i in csv.reader(codecs.iterdecode(zip_city_locations, 'utf-8'), delimiter=',')]
        # field names from 1st row of csv file
        CityLocations = namedtuple('CityLocations', ' '.join(city_locations[0]))
        city_locations = [CityLocations(*x) for x in city_locations[1:]]
        

def get_ip_range(cidr):
    ip_range = iptools.ipv4.cidr2block(cidr)
    return str(ip_range[0]) + '-' + str(ip_range[-1])


def join_data(ip, city):
    city_dict = {x.geoname_id: x for x in city if x.geoname_id is not None}
    join_fields = set(ip[0]._fields + city[0]._fields)
    jd = namedtuple('MMJoinedInfo', ' '.join(join_fields))

    joined_data = []
    for x in ip:
        if x.geoname_id:
            joined_data.append(jd(
                network=x.network,
                geoname_id=x.geoname_id,
                registered_country_geoname_id=x.registered_country_geoname_id,
                represented_country_geoname_id=x.represented_country_geoname_id,
                is_anonymous_proxy=x.is_anonymous_proxy,
                is_satellite_provider=x.is_satellite_provider,
                postal_code=x.postal_code,
                latitude=x.latitude,
                longitude=x.longitude,
                accuracy_radius=x.accuracy_radius,

                # data from city_locations
                locale_code=city_dict[x.geoname_id].locale_code,
                continent_code=city_dict[x.geoname_id].continent_code,
                continent_name=city_dict[x.geoname_id].continent_name,
                country_iso_code=city_dict[x.geoname_id].country_iso_code,
                country_name=city_dict[x.geoname_id].country_name,
                subdivision_1_iso_code=city_dict[x.geoname_id].subdivision_1_iso_code,
                subdivision_1_name=city_dict[x.geoname_id].subdivision_1_name,
                subdivision_2_iso_code=city_dict[x.geoname_id].subdivision_2_iso_code,
                subdivision_2_name=city_dict[x.geoname_id].subdivision_2_name,
                city_name=city_dict[x.geoname_id].city_name,
                metro_code=city_dict[x.geoname_id].metro_code,
                time_zone=city_dict[x.geoname_id].time_zone,
                        ))
    return joined_data


data_set = join_data(ip_blocks, city_locations)

os.makedirs(os.path.dirname(filename), exist_ok=True)

# Creating data for output
# for ipv4
if ipver == 'IPv4':
    if ei_action == 'include':
        output = '\n'.join('{} {};'.format(get_ip_range(data.network),
                                           base64.b64encode(bytes(data.city_name, 'utf-8')).decode('utf-8'))
                           for data in data_set
                           if data.country_iso_code in iso_codes and get_ip_range(data.network) and data.city_name)
    elif ei_action == 'exclude':
        output = '\n'.join('{} {};'.format(get_ip_range(data.network),
                                           base64.b64encode(bytes(data.city_name, 'utf-8')).decode('utf-8'))
                           for data in data_set
                           if data.country_iso_code not in iso_codes and get_ip_range(data.network) and data.city_name)
    else:
        output = '\n'.join('{} {};'.format(get_ip_range(data.network),
                                           base64.b64encode(bytes(data.city_name, 'utf-8')).decode('utf-8'))
                           for data in data_set
                           if get_ip_range(data.network) and data.city_name)
# for ipv6
elif ipver == 'IPv6':
    if ei_action == 'include':
        output = '\n'.join('{} {};'.format(data.network, base64.b64encode(bytes(data.city_name, 'utf-8')).decode('utf-8'))
                           for data in data_set
                           if data.country_iso_code in iso_codes and data.network and data.city_name)
    elif ei_action == 'exclude':
        output = '\n'.join('{} {};'.format(data.network, base64.b64encode(bytes(data.city_name, 'utf-8')).decode('utf-8'))
                           for data in data_set
                           if data.country_iso_code not in iso_codes and data.network and data.city_name)
    else:
        output = '\n'.join('{} {};'.format(data.network, base64.b64encode(bytes(data.city_name, 'utf-8')).decode('utf-8'))
                           for data in data_set
                           if data.network and data.city_name)

with open(filename, 'w') as file:
    file.write(output)
