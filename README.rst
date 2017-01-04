DEPRECATED
==========

Даннная версия импортировщика устаревшая.
Актуальный ip2geo здесь: https://github.com/m-messiah/ip2geo

Please, check out: https://github.com/m-messiah/ip2geo for actual version

IpGeoBase importer
==================

.. image:: https://img.shields.io/pypi/v/ipgeobase-importer.svg?style=flat-square
    :target: https://pypi.python.org/pypi/ipgeobase-importer


.. image:: https://img.shields.io/pypi/dm/ipgeobase-importer.svg?style=flat-square
        :target: https://pypi.python.org/pypi/ipgeobase-importer


.. image:: https://img.shields.io/travis/m-messiah/ipgeobase-importer.svg?style=flat-square
    :target: https://travis-ci.org/m-messiah/ipgeobase-importer

Импортер ipgeobase базы русских городов в файлы, понятные для nginx geoip module, с поддержкой кодов регионов РФ.

Принцип
-------

1.  Скачивает geo_files.zip с сайта ipgeobase.ru
2.  Конвертирует базу в два файла:

    *   city.txt, вида: \<start\_ip\>-\<end\_ip\> base64(\<city_name\>);
    *   region.txt, вида: \<start\_ip\>-\<end\_ip\> \<region\_code\>; (01-99)
3.  Скачивает списки TOR с torproject и blutmagie.de
4.  Создает tor.txt, вида: \<start\_ip\>-\<end\_ip\> 1;

Установка
---------

.. code:: bash

    pip install ipgeobase-importer
    
Запуск
------

.. code:: bash

    ipgeobase-importer <output_dir>
    

Nginx
-----

.. code:: nginx

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


MaxMind geo importer
====================

Скрипт импорта базы MaxMind в map для nginx (спасибо @ilnurkhal)

Требует Python3, requests и iptools.

Формат запуска:

.. code:: bash
    
    ip-maxmind <lang> ipv<ipver> <filename> [filters]

Где:

* ``lang`` - язык, которым будут записаны названия городов (как они указаны в MaxMind)
* ``ipver`` - версия IP для nginx (4 или 6)
* ``filename`` - путь до файла вывода
* filters - фильтры для базы вывода. Могут быть:

    - include <COUNTRY_CODE> [<COUNTRY_CODE>...] - выводить города только из указанных стран.
    - exclude <COUNTRY_CODE> [<COUNTRY_CODE>...] - выводить города, не из указанных стран.
    - или не быть вообще, тогда вывод для всех стран.
