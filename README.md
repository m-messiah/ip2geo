# Ip2Geo importer

[![Travis](https://img.shields.io/travis/m-messiah/ip2geo.svg)](https://travis-ci.org/m-messiah/ip2geo)
[![GitHub release](https://img.shields.io/github/release/m-messiah/ip2geo.svg)](https://github.com/m-messiah/ip2geo)
[![Github Releases](https://img.shields.io/github/downloads/m-messiah/ip2geo/latest/total.svg)](https://github.com/m-messiah/ip2geo)

Импортер ipgeo-данных в файлы, понятные для [nginx geo module](http://nginx.org/ru/docs/http/ngx_http_geo_module.html), с поддержкой кодов регионов РФ.

Поддерживает Ipgeobase.ru, TOR-списки, MaxMind GeoLite (для городов).

## Установка

1. Скачать соответствующий архитектуре бинарник с github куда-нибудь в $PATH
2. Сделать его исполняемым
3. Пользоваться

(также, при наличии Go окружения можно собрать самостоятельно через go get + go build)

## Запуск

По умолчанию, ip2geo генерирует все возможные map-файлы, но все настраиваемо с помощью ключей:

    -output string
        Директория для записи map-файлов (по умолчанию: "output")
    -ipgeobase
        Генерация IPgeobase баз (название города, код региона, часовой пояс)
    -tor
        Генерация списков TOR нод.
    -maxmind
        Генерация баз MaxMind (название города, часовой пояс)
    Дальше параметры для MaxMind:
    -lang string
        Язык MaxMind баз (по умолчанию ru)
    -ipver int
        MaxMind версия IP (4 or 6) (default 4)
    -include string
        MaxMind фильтр: использовать только перечисленные страны  
        Принимает список ISO-кодов стран, разделенных пробелами ("RU FR EN")
    -exclude string
        MaxMind фильтр: исключает из вывода перечисленные страны. (см формат выше)
    

### Формат map-файлов

map-файлы предназначены для использования в nginx в виде:

```nginx
geo $region {
    ranges;
    include geo/region.txt;
}

geo $city {
    ranges;
    include geo/city.txt;
    include geo/mm_city.txt;
}

geo $is_tor {
    ranges;
    default 0;
    include geo/tor.txt;
}

geo $tz {
    ranges;
    default "UTC+3";
    include geo/tz.txt;
    include geo/mm_tz.txt;
}
```

Таким образом, IP адреса в файлах записаны в виде диапазона (range) и отсортированы по возрастанию IP.

Для того, чтобы название города всегда отдавалось корректно - оно кодируется в base64 от utf8.
