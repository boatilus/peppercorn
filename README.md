# peppercorn

**peppercorn** is a simple "post stream"; a single-thread discussion forum concept, written in Go and backed by [RethinkDB](https://www.rethinkdb.com/). Development is early — there are only simple routes defined and models are incomplete.

## Running tests

To run the tests, first [install RethinkDB](https://rethinkdb.com/docs/install/). On macOS, RethinkDB can be installed via Homebrew:

    brew update && brew install rethinkdb
  
Launch a RethinkDB server with default options:

    rethinkdb
  
Run all tests, including subpackages (the `-v` flag specifies verbose output)

    go test -v ./...
