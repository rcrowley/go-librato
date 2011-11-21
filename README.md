go-librato
==========

This is both a Go client to the [Librato Metrics API](http://dev.librato.com/v1/metrics) and a command-line tool for piping data into the client.

Usage
-----

From Go:

```go
m := librato.NewSimpleMetrics(user, token, source)

c := m.GetCounter("foo")
c <- 47

g := m.GetGauge("bar")
c <- 47

cc := m.GetCustomCounter("baz")
cc <- map[string]int64 {
	"value": 47,
	"measure_time": 1234567890,
}

cg := m.GetCustomGauge("bang")
cg <- map[string]int64 {
	"value": 47,
	"measure_time": 1234567890,
}
cg <- map[string]int64 {
	"measure_time": 1234567890,
	"count": 2,
	"sum": 94,
	"max": 47,
	"min": 47,
	"sum_squares": 4418,
}
```

From the command-line:

```sh
thing | librato -u "rcrowley" -t "ZOMG" -s "$(hostname)"

export LIBRATO_USER="rcrowley"
export LIBRATO_TOKEN="ZOMG"
export LIBRATO_SOURCE="$(hostname)"
tail -F /var/log/thing | librato -c 100
```

Installation
------------

Installation requires a working Go build environment.  See their [Getting Started](http://golang.org/doc/install.html) guide if you don't already have one.

As a library:

```sh
goinstall github.com/rcrowley/go-librato
```

The `librato.a` library will by in `$GOROOT/pkg/${GOOS}_${GOARCH}/github.com/rcrowley/go-librato` should be linkable without further configuration.

As a command-line tool:

```sh
git clone git://github.com/rcrowley/go-librato.git
cd go-librato
gomake
```

The `librato` tool will be in `$GOBIN`.
