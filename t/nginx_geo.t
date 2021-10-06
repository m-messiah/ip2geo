use Test::Nginx::Socket 'no_plan';
run_tests();

__DATA__

=== TEST 1: ip2geo
--- main_config
    error_log  $TEST_NGINX_IP2GEO_DIR/error.log;
--- http_config

# City
    geo $city {
        ranges;
        include $TEST_NGINX_IP2GEO_DIR/output/mm_city.txt;
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
    geo $tz {
        ranges;
        include $TEST_NGINX_IP2GEO_DIR/output/mm_tz.txt;
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
