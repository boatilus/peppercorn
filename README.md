# peppercorn

[![Go Report Card](https://goreportcard.com/badge/github.com/boatilus/peppercorn)](https://goreportcard.com/report/github.com/boatilus/peppercorn) [![Build Status](https://travis-ci.org/boatilus/peppercorn.svg?branch=master)](https://travis-ci.org/boatilus/peppercorn)

**peppercorn** is a simple "post stream"; a private, single-thread discussion forum concept, written in Go and backed by [RethinkDB](https://www.rethinkdb.com/). Development is early — there are only simple routes defined and models are incomplete.

## Installing

If you've [installed RethinkDB](https://rethinkdb.com/docs/install/), you're ready to `go get`:

    go get -u github.com/boatilus/peppercorn
    
**peppercorn** will create the database and tables necessary for itself to function — if they don't already exist — each time it's started.

## Running tests

To run the tests, first [install RethinkDB](https://rethinkdb.com/docs/install/). On macOS, RethinkDB can be installed via Homebrew:

    brew update && brew install rethinkdb
  
Launch a RethinkDB server with default options:

    rethinkdb
  
Run all tests, including subpackages (the optional `-v` flag specifies verbose output)

    go test -v ./...

## Configuration

**peppercorn** uses the most-excellent [Viper](https://github.com/spf13/viper) for configuration, so you can supply **peppercorn** with either a JSON or a YAML `config` file. At this time, this file won't be auto-created with defaults on first run, so it needs to be done manually.

Your `config` should look like this and be placed in the app's working directory:

    {
      "test": true,
      "port": ":8000",
      "bcrypt_cost": 10,
      "session_key": "anyString",
      "cookie": {
        "hash_key": "a 64-character string for HMAC",
        "block_key": "a 32-character string for AES-256"
      },
      "db": {
        "address": "localhost:28015",
        "name": "peppercorn",
        "users_table": "users",
        "posts_table": "posts",
        "sessions_table": "sessions"
      },
      "sentry": {
        "dsn": "your Sentry DSN, if desired"
      }
    }
    
Remove or change `test` to false for deployment to production and modify `bcrypt_cost` to suit your specific security needs and your runtime environment. [This article by Joseph Wynn](https://wildlyinaccurate.com/bcrypt-choosing-a-work-factor/) explains how one might go about choosing a suitable cost (work factor).
