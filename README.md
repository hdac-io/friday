[![Travis](https://travis-ci.com/hdac-io/friday.svg?token=bhU3g7FdixBp5h3M2its&branch=dev)](https://travis-ci.com/hdac-io/friday/branches)
[![codecov](https://codecov.io/gh/hdac-io/friday/branch/dev/graph/badge.svg?token=hQEgzmULjh)](https://codecov.io/gh/hdac-io/friday))

# TESTNET

* go 1.13 is required

* Step to install
    * change directory to project root dir and run command below
    * `$ go build ./cmd/fryd && go build ./cmd/friday`

* How to run
    * Daemon: `$ ./fryd`
    * CLI: `$ ./friday`

* How to test
    * `` $ for xx in `find ./ | grep _test.go | awk -F/ '{for(i=1;i<NF;i++) printf $i"/"; print ""}' | sort | uniq`; do; go test $xx; done ``
