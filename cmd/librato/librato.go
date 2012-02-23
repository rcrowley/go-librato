package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/rcrowley/go-librato"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
)

// Storage for flags.
var collate int
var user, token, source string

// Create a map suitable for use as a custom counter metric from the given
// regular expression match.
func customCounter(match []string) map[string]int64 {
	obj := make(map[string]int64)
	value, err := strconv.ParseInt(match[3], 10, 64)
	if nil == err {
		obj["value"] = value
	}
	measureTime, err := strconv.ParseInt(match[4], 10, 64)
	if nil == err {
		obj["measure_time"] = measureTime
	}
	return obj
}

// Create a map suitable for use as a custom gauge metric from the given
// regular expression match.
func customGauge(match []string) map[string]int64 {
	obj := make(map[string]int64)
	value, err := strconv.ParseInt(match[3], 10, 64)
	if nil == err {
		obj["value"] = value
	}
	measureTime, err := strconv.ParseInt(match[4], 10, 64)
	if nil == err {
		obj["measure_time"] = measureTime
	}
	count, err := strconv.ParseInt(match[5], 10, 64)
	if nil == err {
		obj["count"] = count
	}
	sum, err := strconv.ParseInt(match[6], 10, 64)
	if nil == err {
		obj["sum"] = sum
	}
	max, err := strconv.ParseInt(match[7], 10, 64)
	if nil == err {
		obj["max"] = max
	}
	min, err := strconv.ParseInt(match[8], 10, 64)
	if nil == err {
		obj["min"] = min
	}
	sumSquares, err := strconv.ParseInt(match[9], 10, 64)
	if nil == err {
		obj["sum_squares"] = sumSquares
	}
	return obj
}

// Initialize the flags with their default values from the environment.
func init() {
	flag.IntVar(
		&collate,
		"c",
		0,
		"maximum number of Librato Metrics API requests to collate",
	)
	flag.StringVar(
		&user,
		"u",
		os.Getenv("LIBRATO_USER"),
		"Librato user",
	)
	flag.StringVar(
		&token,
		"t",
		os.Getenv("LIBRATO_TOKEN"),
		"Librato API token",
	)
	flag.StringVar(
		&source,
		"s",
		os.Getenv("LIBRATO_SOURCE"),
		"metric source",
	)
}

func main() {

	log.SetFlags(log.Ltime | log.Lmicroseconds | log.Lshortfile)

	flag.Usage = usage
	flag.Parse()

	// The `user` and `token` flags are required.  The `source` flag is not.
	if "" == user {
		log.Fatalln("no Librato user found in -u or LIBRATO_USER")
	}
	if "" == token {
		log.Fatalln("no Librato API token found in -t or LIBRATO_TOKEN")
	}

	// Create a Librato Metrics client with the given credentials and source.
	var m librato.Metrics
	if 0 < collate {
		m = librato.NewCollatedMetrics(user, token, source, collate)
	} else {
		m = librato.NewSimpleMetrics(user, token, source)
	}

	// Regular expressions for parsing standard input.  Valid lines contain
	// a literal 'c' or 'g' character to identify the type of the metric, a
	// name, and one or more numeric fields.  Counters can accomodate up to
	// two numeric fields, with the second representing the `measure_time`
	// field.  Gauges can accommodate up to seven numeric fields, which
	// represent, in order, `value`, `measure_time`, `count`, `sum`, `max`,
	// `min`, and `sum_squares` as documented by Librato:
	// <http://dev.librato.com/v1/post/gauges/:name>.
	//
	//                           cg    name    value
	re := regexp.MustCompile("^([cg]) ([^ ]+) ([0-9]+)$")
	//                                       c   name    value  measure_time
	reCustomCounter := regexp.MustCompile("^(c) ([^ ]+) ([0-9]+) (-|[0-9]+)$")
	reCustomGauge := regexp.MustCompile(
		//     g   name      value  measure_time   count      sum
		"^(g) ([^ ]+) (-|[0-9]+) (-|[0-9]+) (-|[0-9]+) (-|[0-9]+) " +
			//      max        min     sum_squares
			"(-|[0-9]+) (-|[0-9]+) (-|[0-9]+)$")

	// Read standard input line-buffered.  Break out of this loop on EOF.
	// Log an error message and exit if any other error is encountered.
	stdin := bufio.NewReader(os.Stdin)
	for {
		line, _, err := stdin.ReadLine()
		s := string(line)
		if io.EOF == err {
			break
		}
		if nil != err {
			log.Fatalln(err)
		}

		// Match this line against the regular expressions above.  In
		// case a line doesn't match, log the line and continue.  Get
		// the appropriate channel and send the metric.
		if match := re.FindStringSubmatch(s); nil != match {
			var ch chan int64
			switch match[1] {
			case "c":
				ch = m.GetCounter(match[2])
			case "g":
				ch = m.GetGauge(match[2])
			}
			value, _ := strconv.ParseInt(match[3], 10, 64)
			ch <- value
		} else if match := reCustomCounter.FindStringSubmatch(s); nil != match {
			m.GetCustomCounter(match[2]) <- customCounter(match)
		} else if match := reCustomGauge.FindStringSubmatch(s); nil != match {
			m.GetCustomGauge(match[2]) <- customGauge(match)
		} else {
			log.Printf("malformed line \"%v\"\n", s)
		}

	}

	// Close all metric channels so no new messages may be sent.  Wait
	// for all outstanding HTTP requests to finish.
	//
	// This can deadlock in the event that EOF is seen (and hence execution
	// arrives here) before a single metric has been sent.  The Go runtime
	// will detect the deadlock and abort with a nasty stack trace.
	m.Close()
	m.Wait()

}

func usage() {
	fmt.Fprintln(
		os.Stderr,
		"Usage: librato [-c <collate>] [-u <user>] [-t <token>] [-s <source>]",
	)
}
