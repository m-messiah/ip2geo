use Test::Nginx::Socket 'no_plan';
run_tests();

__DATA__

=== TEST 1: ip2geo
--- http_config

geo $region {
    ranges;
    include $TEST_NGINX_IP2GEO_DIR/output/region.txt;
}

geo $city {
    ranges;
    default $city_mm;
    include $TEST_NGINX_IP2GEO_DIR/output/city.txt;
}

geo $city_mm {
    ranges;
    include $TEST_NGINX_IP2GEO_DIR/output/mm_city.txt;
}

geo $is_tor {
    ranges;
    default 0;
    include $TEST_NGINX_IP2GEO_DIR/output/tor.txt;
}

geo $tz {
    ranges;
    default $tz_mm;
    include $TEST_NGINX_IP2GEO_DIR/output/tz.txt;
}

geo $tz_mm {
    ranges;
    default "UTC+3";
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
