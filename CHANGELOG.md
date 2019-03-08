v1.0.6 / 2019-03-08
==================

  * Add country generation from MaxMind
  * Add county iso_code generation from MaxMind
  * Add `-nocountry` flag to skip MaxMind country and country code generation
  * Fix IPv6 generation for MaxMind

v1.0.5 / 2019-02-19
==================

  * Add `-nobase64` flag to skip base64 MaxMind cities encoding
  * Add multiarch build/release

v1.0.4 / 2018-03-01
==================

  * Add ip2proxy (www.ip2location.com) support (fix #9) (@AlexAkulov)
  * Update example in README to correct geo maps usage (close #7)
  * Add nginx geo maps test case (close #4)

v1.0.3 / 2017-07-20
===================

  * Fix tor merge, when one is nil (close #6)

v1.0.2 / 2017-07-03
===================

  * Add `-q` and `-qq` params to reduce output
  * Add error handling (fix #5)

v1.0.1 / 2017-06-08
===================

  * Repair travis deploy
  * Move old version to branch
  * Add MaxMind TZ in offsets format (names format is possible too)
  * Refactor for GoLint
  * Add geo module docs link
  * Add releases downloads badge
  
v1.0.0 / 2017-01-04
===================
First Golang release
  * Speedup sorting form MaxMind
  * Add Github releases
  * Rename to ip2geo
