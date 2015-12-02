#!/bin/sh -e

rm -rf /usr/local/etc/geo.bak
cp -r /usr/local/etc/geo /usr/local/etc/geo.bak || echo "geo not exists"

/usr/local/bin/ipgeobase-importer /usr/local/etc/geo/
/usr/local/sbin/nginx -t || (echo "Fail" >&2 && cp -r /usr/local/etc/geo.bak/* /usr/local/etc/geo/ && exit 1)

