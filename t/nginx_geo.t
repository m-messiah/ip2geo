use Test::Nginx::Socket 'no_plan';
run_tests();

__DATA__

=== TEST 1: ip2geo
--- http_config

# Region
    geo $region {
        ranges;
        include $TEST_NGINX_IP2GEO_DIR/output/region.txt;
    }
# City
    geo $city_geo {
        ranges;
        include $TEST_NGINX_IP2GEO_DIR/output/city.txt;
    }

    geo $city_mm {
        ranges;
        include $TEST_NGINX_IP2GEO_DIR/output/mm_city.txt;
    }

    map $city_geo $city {
        "" $city_mm;
        default $city_geo;
    }
# Country
    geo $country {
        ranges;
        include $TEST_NGINX_IP2GEO_DIR/output/mm_country.txt;
    }
# CountryCode
    geo $country_code {
        ranges;
        include $TEST_NGINX_IP2GEO_DIR/output/mm_country_code.txt;
    }
# TZ
    geo $tz_geo {
        ranges;
        include $TEST_NGINX_IP2GEO_DIR/output/tz.txt;
    }

    geo $tz_mm {
        ranges;
        include $TEST_NGINX_IP2GEO_DIR/output/mm_tz.txt;
    }

    map $tz_geo $tz {
        "" $tz_mm;
        default $tz_geo;
    }
# Tor
    geo $is_tor {
        ranges;
        default 0;
        include $TEST_NGINX_IP2GEO_DIR/output/tor.txt;
    }

--- config
    location /t {
        default_type text/plain;
        return 200 "Ok";
    }
--- request
GET /t
--- error_code: 200
--- no_error_log
[error]
