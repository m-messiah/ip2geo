use Test::Nginx::Socket 'no_plan';
run_tests();

__DATA__

=== TEST 1: ip2geo -ipv6
--- main_config
    error_log  $TEST_NGINX_IP2GEO_DIR/error.log;
--- http_config

# City
    geo $city {
        include $TEST_NGINX_IP2GEO_DIR/output/mm_city.txt;
    }

# Country
    geo $country {
        include $TEST_NGINX_IP2GEO_DIR/output/mm_country.txt;
    }
# CountryCode
    geo $country_code {
        include $TEST_NGINX_IP2GEO_DIR/output/mm_country_code.txt;
    }
# TZ
    geo $tz {
        include $TEST_NGINX_IP2GEO_DIR/output/mm_tz.txt;
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
