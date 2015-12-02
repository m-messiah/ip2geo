#!/bin/sh -e

cd /usr/local/etc
rm -rf geo.bak
cp -r geo geo.bak

/usr/local/bin/python /root/ipgeobase-importer/ipgeobase-importer.py
/usr/local/sbin/nginx -t || (echo "Fail" && cp -r geo.bak/* geo/ && exit 1)

