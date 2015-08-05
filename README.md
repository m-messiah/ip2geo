# Geo IP importer

Импортер geoIP базы русских городов в файлы, понятные для nginx geoip module, с поддержкой кодов регионов РФ.

## Принцип

1.  Скачивает geo_files.zip с сайта ipgeobase.ru
2.  Конвертирует базу в два файла:
    +   city.txt, вида: <start_ip>-<end_ip> base64(<city_name>);
    +   region.txt, вида: <start_ip>-<end_ip> <region_code>;

## Зависимости

+   Python 2.7
+   python-requests (pip install requests)

## Nginx

```
geo $http_x_ip $region {
ranges;
default 0;
include geo/region.txt;
}

geo $http_x_ip $city {
ranges;
default 0;
include geo/city.txt;
}
```
