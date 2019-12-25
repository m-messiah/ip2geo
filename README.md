# Ip2Geo importer

[![Travis](https://img.shields.io/travis/m-messiah/ip2geo.svg?style=for-the-badge)](https://travis-ci.org/m-messiah/ip2geo)[![GitHub release](https://img.shields.io/github/release/m-messiah/ip2geo.svg?style=for-the-badge)](https://github.com/m-messiah/ip2geo/releases/latest)[![Github Releases](https://img.shields.io/github/downloads/m-messiah/ip2geo/total.svg?style=for-the-badge)](https://github.com/m-messiah/ip2geo/releases)[![Maintainability](https://api.codeclimate.com/v1/badges/8687e760d260b8499393/maintainability)](https://codeclimate.com/github/m-messiah/ip2geo/maintainability)

Импортер ipgeo-данных в файлы, понятные для [nginx geo module](http://nginx.org/ru/docs/http/ngx_http_geo_module.html), с поддержкой кодов регионов РФ.

Поддерживает Ipgeobase.ru, TOR-списки, MaxMind GeoLite (для городов и стран), базы IP2Proxy PX4 и PX4LITE.

## Установка

1. Скачать соответствующий архитектуре бинарник с github куда-нибудь в $PATH
2. Сделать его исполняемым
3. Пользоваться

(также, при наличии Go окружения можно собрать самостоятельно через go get + go build)

## Запуск

По умолчанию, ip2geo генерирует все возможные map-файлы, но все настраиваемо с помощью ключей:

    -c      string
        Путь до конфигурационного файла (см. config.yaml.example)
    -output string
        Директория для записи map-файлов (по умолчанию: "output")
    -q  Be quiet - skip [OK]
    -qq Be very quiet - show only errors
    -version Вывести текущую версию программы и выйти.
    -ipgeobase
        Генерация IPgeobase баз (название города, код региона, часовой пояс)
    -tor
        Генерация списков TOR нод.
    -ip2proxy
        Генерация ip2proxy PX4LITE сетей
    -ip2proxy-token string
        Токен для скачивания ip2proxy PX4LITE баз https://lite.ip2location.com/file-download
    -ip2proxy-lite-filename string
        Путь до уже скачанного zip файла PX4LITE
    -ip2proxy-pro
        Генерация ip2proxy PX4 сетей
    -ip2proxy-pro-token string
        Токен для скачивания ip2proxy PX4
    -ip2proxy-pro-filename string
        Путь до уже скачанного zip файла PX4
    -ip2proxy-print-type
        Вместо 1 указывать тип прокси из ip2proxy (PUB/DCH/e.t.c.)
    -maxmind
        Генерация баз MaxMind (название города, часовой пояс)
    Дальше параметры для MaxMind:
    -maxmind-license-key string
        Лицензионный ключ для MaxMind
    -maxmind-filename string
        Путь до уже скачанного файла MaxMind GeoLite2-City-CSV.csv
    -lang string
        Язык MaxMind баз (по умолчанию ru)
    -ipver int
        MaxMind версия IP (4 or 6) (default 4)
    -include string
        MaxMind фильтр: использовать только перечисленные страны  
        Принимает список ISO-кодов стран, разделенных пробелами ("RU FR EN")
    -exclude string
        MaxMind фильтр: исключает из вывода перечисленные страны. (см формат выше)
    -nobase64
        Не перекодирует MaxMind города в base64, записывая их в map-файл как есть. Не используйте, если не уверены в кодировке MaxMind.
    -nocountry
        Не создавать map-файлы с названиями стран и iso-кодами стран из MaxMind
    

### Формат geomap-файлов

geomap-файлы предназначены для использования в nginx в виде:

```nginx
# Region
    geo $region {
        ranges;
        include geo/region.txt;
    }
# City
    geo $city_geo {
        ranges;
        include geo/city.txt;
    }

    geo $city_mm {
        ranges;
        include geo/mm_city.txt;
    }

    map $city_geo $city {
        "" $city_mm;
        default $city_geo;
    }
# Country
    geo $country {
        ranges;
        include geo/mm_country.txt;
    }
# Country Code
    geo $country_code {
        ranges;
        include geo/mm_country_code.txt;
    }
# TZ
    geo $tz_geo {
        ranges;
        include geo/tz.txt;
    }

    geo $tz_mm {
        ranges;
        include geo/mm_tz.txt;
    }

    map $tz_geo $tz {
        "" $tz_mm;
        default $tz_geo;
    }
# Tor
    geo $is_tor {
        ranges;
        default 0;
        include geo/tor.txt;
    }
# Proxy
    geo $is_proxy {
        ranges;
        default 0;
        include geo/ip2proxy_net.txt;
    }
```

Таким образом, IP адреса в файлах записаны в виде диапазона (range) и отсортированы по возрастанию IP. Карты сделаны каскадно, чтобы решить проблему пересечений диапазонов. IPGeobase используется в первую очередь, и если адрес там не найден, то MaxMind.

Для того, чтобы название города всегда отдавалось корректно - оно кодируется в base64 от utf8 (если не указан флаг `-nobase64`).
