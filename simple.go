package librato

import (
	"bytes"
	"fmt"
	"http"
	"json"
	"log"
	"os"
	"url"
)

// Librato `SimpleMetrics` structs encapsulate the credentials used to send
// metrics to the API, the source tag for these metrics, bookkeeping for
// goroutines, and lookup tables for existing metric channels.
type SimpleMetrics struct {
	user, token, source string
	quit, running chan bool
	counters, gauges map[string]chan int64
	customCounters, customGauges map[string]chan map[string]int64
}

// Create a new `SimpleMetrics` struct with the given credentials and source
// tag.  Initialize all the channels, maps, and goroutines used internally.
func NewSimpleMetrics(user, token, source string) Metrics {
	m := &SimpleMetrics{
		user, token, source,
		make(chan bool), make(chan bool),
		make(map[string]chan int64), make(map[string]chan int64),
		make(map[string]chan map[string]int64),
		make(map[string]chan map[string]int64),
	}

	// Track the number of running goroutines.  When it returns to zero,
	// send a message to the quit channel.
	go func() {
		var n uint
		for {
			if <-m.running { n++ } else if 0 < n { n-- }
			if 0 == n { break }
		}
		m.quit <- true
	}()

	return m
}
func NewMetrics(user, token, source string) Metrics {
	return NewSimpleMetrics(user, token, source)
}

// Close all metric channels so no new messages may be sent.  This is
// a prerequisite to `Wait`ing.
func (m *SimpleMetrics) Close() {
	for _, ch := range m.counters { close(ch) }
	for _, ch := range m.gauges { close(ch) }
	for _, ch := range m.customCounters { close(ch) }
	for _, ch := range m.customGauges { close(ch) }
}

// Get (possibly by creating) a counter channel by the given name.
func (m *SimpleMetrics) GetCounter(name string) chan int64 {
	ch, ok := m.counters[name]
	if ok { return ch }
	return m.NewCounter(name)
}

// Get (possibly by creating) a custom counter channel by the given name.
func (m *SimpleMetrics) GetCustomCounter(name string) chan map[string]int64 {
	ch, ok := m.customCounters[name]
	if ok { return ch }
	return m.NewCustomCounter(name)
}

// Get (possibly by creating) a custom gauge channel by the given name.
func (m *SimpleMetrics) GetCustomGauge(name string) chan map[string]int64 {
	ch, ok := m.customGauges[name]
	if ok { return ch }
	return m.NewCustomGauge(name)
}

// Get (possibly by creating) a gauge channel by the given name.
func (m *SimpleMetrics) GetGauge(name string) chan int64 {
	ch, ok := m.gauges[name]
	if ok { return ch }
	return m.NewGauge(name)
}

// Create a counter channel by the given name.
func (m *SimpleMetrics) NewCounter(name string) chan int64 {
	ch := m.newMetric("/counters/%s.json", name)
	m.counters[name] = ch
	return ch
}

// Create a custom counter channel by the given name.
func (m *SimpleMetrics) NewCustomCounter(name string) chan map[string]int64 {
	ch := m.newCustomMetric("/counters/%s.json", name)
	m.customCounters[name] = ch
	return ch
}

// Create a custom gauge channel by the given name.
func (m *SimpleMetrics) NewCustomGauge(name string) chan map[string]int64 {
	ch := m.newCustomMetric("/gauges/%s.json", name)
	m.customGauges[name] = ch
	return ch
}

// Create a gauge channel by the given name.
func (m *SimpleMetrics) NewGauge(name string) chan int64 {
	ch := m.newMetric("/gauges/%s.json", name)
	m.gauges[name] = ch
	return ch
}

// Wait for all outstanding HTTP requests to finish.  This must be called
// after `Close` has been called.
func (m *SimpleMetrics) Wait() {
	<-m.quit
}

// Serialize an `application/json` request body and do one HTTP roundtrip
// using the `http` package's `DefaultClient`.  This wrapper constructs the
// appropriate Librato Metrics API endpoint, sets the `Content-Type` header
// to `application/json`, and sets the `Authorization` header for  HTTP Basic
// authentication from the `SimpleMetrics` struct.
func (m *SimpleMetrics) do(
	format, name string,
	body map[string]interface{},
) os.Error {
	if "" != m.source { body["source"] = m.source }
	b, err := json.Marshal(body)
	if nil != err { return err }
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"https://metrics-api.librato.com/v1%s",
			fmt.Sprintf(
				format,
				url.QueryEscape(name),
			),
		),
		bytes.NewBuffer(b),
	)
	if nil != err { return err }
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(m.user, m.token)
	fmt.Printf("req: %v\n", req)
	resp, err := http.DefaultClient.Do(req)
	fmt.Printf("resp: %v\n", resp)
	return err
}

// Create a custom metric channel and begin processing messages sent
// to it in a background goroutine.
func (m *SimpleMetrics) newCustomMetric(
	format, name string,
) chan map[string]int64 {
	ch := make(chan map[string]int64)
	go func() {
		m.running <- true
		for {
			obj, ok := <-ch
			if !ok { break }
			body := make(map[string]interface{})
			for k, v := range obj { body[k] = v }
			err := m.do(format, name, body)
			if nil != err { log.Println(err) }
		}
		m.running <- false
	}()
	return ch
}

// Create a metric channel and begin processing messages sent
// to it in a background goroutine.
func (m *SimpleMetrics) newMetric(format, name string) chan int64 {
	ch := make(chan int64)
	go func() {
		m.running <- true
		for {
			v, ok := <-ch
			if !ok { break }
			body := map[string]interface{} { "value": v }
			err := m.do(format, name, body)
			if nil != err { log.Println(err) }
		}
		m.running <- false
	}()
	return ch
}
