# Single Use Dialer

[![GoDoc](https://pkg.go.dev/badge/github.com/bassosimone/sud)](https://pkg.go.dev/github.com/bassosimone/sud) [![Build Status](https://github.com/bassosimone/sud/actions/workflows/go.yml/badge.svg)](https://github.com/bassosimone/sud/actions) [![codecov](https://codecov.io/gh/bassosimone/sud/branch/main/graph/badge.svg)](https://codecov.io/gh/bassosimone/sud)

The `sud` Go package provides Single Use Dialers.

A single use dialer allows injecting a pre-established `net.Conn`` into
components that expect to control dialing themselves, such as `http.Transport`.
The first dial succeeds and returns the injected connection; subsequent
dials fail with `ErrNoConnReuse`.

The name "sud" is an acronym for "Single Use Dialer" (and also means
"south" in Italian, which is a nice coincidence).

## Usage

```Go
import "github.com/bassosimone/sud"

// Create a dialer wrapping an existing connection
conn, _ := net.Dial("tcp", "example.com:443")
dialer := sud.NewSingleUseDialer(conn)

// Use with http.Transport
transport := &http.Transport{
    DialContext: dialer.DialContext,
}
```

## Installation

To add this package as a dependency to your module:

```sh
go get github.com/bassosimone/sud
```

## Development

To run the tests:
```sh
go test -v .
```

To measure test coverage:
```sh
go test -v -cover .
```

## License

```
SPDX-License-Identifier: GPL-3.0-or-later
```

## History

Adapted from [ooni/probe-cli](https://github.com/ooni/probe-cli/blob/v3.20.1/internal/netxlite/dialer.go).
