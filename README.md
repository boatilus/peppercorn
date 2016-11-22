# peppercorn

**peppercorn** is a simple "post stream"; a single-thread discussion forum concept, written in Go and backed by RethinkDB. Development is early — there are only simple routes defined and models are incomplete.

## Testing

To run the tests, first [install RethinkDB](https://rethinkdb.com/docs/install/). On macOS, RethinkDB can be installed via Homebrew:

    brew update && brew install rethinkdb
  
Launch the RethinkDB server with default options:

    rethinkdb
  
Currently tests are available for the `users` and `posts` modules.

    cd users && go test -v
    cd posts && go test- v
