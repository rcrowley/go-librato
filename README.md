go-librato
==========

This is both a Go client to the [Librato Metrics API](http://dev.librato.com/v1/metrics) and a command-line tool for piping data into the client.

Usage
-----

From Go:

```go
m := librato.NewSimpleMetrics(user, token, source)
defer m.Wait()
defer m.Close()

c := m.GetCounter("foo")
c <- 47

g := m.GetGauge("bar")
g <- 47

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

Alternatively you can use the collated mode so data will be sent when
enough measurements are available:

```go
collate_max := 3
m := librato.NewCollatedMetrics(user, token, source, collate_max)
c := m.GetCounter("foo")
g := m.GetGauge("bar")
c <- 1
c <- 2
c <- 3 // send here ...
g <- 10
g <- 20
g <- 30 // send here ...
```

As above, custom metrics are also available.

You can also use the command line tool to send your data:

```sh
thing | librato -u "rcrowley" -t "ZOMG" -s "$(hostname)"

export LIBRATO_USER="rcrowley"
export LIBRATO_TOKEN="ZOMG"
export LIBRATO_SOURCE="$(hostname)"
tail -F /var/log/thing | librato -c 100
```

The `librato` tool accepts one metric per line.  The first field is either a `c` or a `g` to indicate that the metric is a counter or a gauge.  The second field is the name of the metric, which may not contain spaces.  The remaining fields may either be numeric or `-` but must provide a combination of non-`-` values acceptable to the Librato Metrics API.

Regular expressions:

```
# Value-only counters and gauges.
^([cg]) ([^ ]+) ([0-9]+)$

# Custom counters with a value and optionally a timestamp.
^(c) ([^ ]+) ([0-9]+) (-|[0-9]+)$

# Custom gauges with a value, timestamp, count, sum, max, min, and sum-of-squares (or some combination thereof).
^(g) ([^ ]+) (-|[0-9]+) (-|[0-9]+) (-|[0-9]+) (-|[0-9]+) (-|[0-9]+) (-|[0-9]+) (-|[0-9]+)$
```

Examples:

```
c foo 47
g bar 47
c baz 47 1234567890
g bang 47 1234567890 - - - - -
g bang - 1234567890 2 94 47 47 4418
```

Installation
------------

Installation requires a working Go build environment.  See their [Getting Started](http://golang.org/doc/install.html) guide if you don't already have one.

As a library:

```sh
go get github.com/rcrowley/go-librato
```

The `librato.a` library will by in `$GOROOT/pkg/${GOOS}_${GOARCH}/github.com/rcrowley/go-librato` should be linkable without further configuration.

As a command-line tool:

```sh
git clone git://github.com/rcrowley/go-librato.git
cd go-librato/cmd/librato
go install
```

The `librato` tool will be in `$GOBIN`.
