#!/bin/sh -e

export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
export GEO_LOCATION=/usr/local/etc/geo

rm -rf ${GEO_LOCATION}.bak
cp -r ${GEO_LOCATION} ${GEO_LOCATION}.bak || echo "geo not exists"

/usr/local/bin/ipgeobase-importer ${GEO_LOCATION} 2> /var/log/ipgeobase.log
nginx -t || (echo "Fail" >&2 && cp -r ${GEO_LOCATION}.bak/* ${GEO_LOCATION} && exit 1)

