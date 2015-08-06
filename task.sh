#!/bin/bash -e

pushd /usr/local/etc/nginx
rm -rf geo.bak
mv geo geo.bak
mkdir geo

python ipgeobase_importer.py
nginx -t || (echo "Fail" && rm -rf geo && mv geo.bak geo && exit 1)

popd
