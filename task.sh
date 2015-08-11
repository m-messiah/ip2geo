#!/bin/sh -e

cd /usr/local/etc
rm -rf geo.bak
cp -r geo geo.bak

ipgeobase_importer
nginx -t || (echo "Fail" && cp -r geo.bak/* geo/ && exit 1)

