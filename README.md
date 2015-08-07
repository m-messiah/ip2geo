# IpGeoBase importer

Импортер ipgeobase базы русских городов в файлы, понятные для nginx geoip module, с поддержкой кодов регионов РФ.

## Принцип

1.  Скачивает geo_files.zip с сайта ipgeobase.ru
2.  Конвертирует базу в два файла:
    +   city.txt, вида: \<start\_ip\>-\<end\_ip\> base64(\<city_name\>);
    +   region.txt, вида: \<start\_ip\>-\<end\_ip\> \<region\_code\>;

## Зависимости

+   Python 2.7
+   python-requests (`pip install requests`)

## Nginx

```nginx
geo $region {
    ranges;
    include geo/region.txt;
}

geo $city {
    ranges;
    include geo/city.txt;
}

geo $is_tor {
    ranges;
    default 0;
    include geo/tor.txt;
}
```
