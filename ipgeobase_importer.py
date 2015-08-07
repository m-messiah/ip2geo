#!/usr/bin/env python

from requests import get as rget
from zipfile import ZipFile
from StringIO import StringIO
from base64 import b64encode
from os import path as os_path


def error(text):
    sys.stderr.write(text + "\n")
    exit(1)

archive = rget("http://ipgeobase.ru/files/db/Main/geo_files.zip")
if archive.status_code != 200:
    error("IPGeobase no answer: %s" % archive.status_code)

extracteddata = ZipFile(StringIO(archive.content))

filelist = extracteddata.namelist()
if "cities.txt" not in filelist:
    error("cities.txt not downloaded")
if "cidr_optim.txt" not in filelist:
    error("cidr_optim.txt not downloaded")

database = {}

REGIONS = dict(l.decode("utf8").rstrip().split("\t")[::-1]
               for l in open(
                   os_path.dirname(os_path.realpath(__file__)) +
                   "/regions.tsv").readlines())
CITIES = {}
for line in extracteddata.open("cities.txt").readlines():
    # Format is:
    # <city_id>\t<city_name>\t<region>\t<district>\t<lattitude>\t<longitude>
    cid, city, region_name, _, _, _ = line.decode("cp1251").split("\t")
    if region_name in REGIONS:
        CITIES[cid] = {'city': b64encode(city.encode("utf8")),
                       'reg_id': REGIONS[region_name]}
        if cid == "1199":  # Zelenograd fix
            CITIES[cid]['reg_id'] = "77"

for line in extracteddata.open("cidr_optim.txt").readlines():
    # Format is: <int_start>\t<int_end>\t<ip_range>\t<country_code>\tcity_id
    _, _, ip_range, country, cid = line.decode("cp1251").rstrip().split("\t")
    # Skip not russian cities
    if country == "RU" and cid in CITIES:
            database["".join(ip_range.split())] = CITIES[cid]


# Create nginx geoip compatible files
with open("geo/region.txt", "w") as reg, open("geo/city.txt", "w") as city:
    for ip_range in sorted(database):
        info = database[ip_range]
        city.write("%s %s;\n" % (ip_range, info['city']))
        reg.write("%s %s;\n" % (ip_range, info['reg_id']))

# Tor
torlist = rget("https://torstatus.blutmagie.de/ip_list_exit.php/Tor_ip_list_EXIT.csv")
if torlist.status_code == 200:
    torlist = set(filter(len, torlist.content.split("\n")))
else:
    torlist = set()

torproject = rget("https://check.torproject.org/exit-addresses")
if torproject.status_code == 200:
    torproject = set(
                    map(lambda s: s.split()[1],
                        filter(lambda l: "ExitAddress" in l,
                               torproject.content.split("\n"))))
else:
    torproject = set()

with open("geo/tor.txt", "w") as tor:
    for ip in sorted(torlist|torproject, key=lambda i: tuple(int(p) for p in i.split("."))):
        tor.write("%s-%s 1;\n" % (ip, ip))

