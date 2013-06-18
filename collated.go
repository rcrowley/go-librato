package librato

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Librato `CollatedMetrics` structs encapsulate the credentials used to send
// metrics to the API, the source tag for these metrics, the maximum number
// of metrics to collate, bookkeeping for goroutines, and lookup tables for
// existing metric channels.
type CollatedMetrics struct {
	user, token, source            string
	collateMax                     int
	quit, running                  chan bool
	collateCounters, collateGauges chan map[string]interface{}
	counters, gauges               map[string]chan int64
	customCounters, customGauges   map[string]chan map[string]int64
}

// Create a new `CollatedMetrics` struct with the given credentials, source
// tag, and maximum collation.  Initialize all the channels, maps, and
// goroutines used internally.
func NewCollatedMetrics(user, token, source string,
	collateMax int) Metrics {
	m := &CollatedMetrics{
		user, token, source,
		collateMax,
		make(chan bool), make(chan bool),
		make(chan map[string]interface{}, collateMax),
		make(chan map[string]interface{}, collateMax),
		make(map[string]chan int64), make(map[string]chan int64),
		make(map[string]chan map[string]int64),
		make(map[string]chan map[string]int64),
	}

	// Track the number of running goroutines.  When it returns to zero,
	// close the collation channels.
	go func() {
		var n uint
		for {
			if <-m.running {
				n++
			} else if 0 < n {
				n--
			}
			if 0 == n {
				break
			}
		}
		close(m.collateCounters)
		close(m.collateGauges)
	}()

	// Receive metric bodies on the collation channels.
	go func() {
		for {
			i := 0
			ok := true
			counters := make([]map[string]interface{}, 0, m.collateMax)
			gauges := make([]map[string]interface{}, 0, m.collateMax)
			for i < m.collateMax {
				var body map[string]interface{}
				select {
				case body, ok = <-m.collateCounters:
					if ok {
						counters = append(counters, body)
					}
				case body, ok = <-m.collateGauges:
					if ok {
						gauges = append(gauges, body)
					}
				}
				if ok {
					i++
				} else {
					break
				}
			}
			if 0 < i {
				err := m.do(map[string]interface{}{
					"counters": counters,
					"gauges":   gauges,
				})
				if nil != err {
					log.Println(err)
				}
			}
			if !ok {
				break
			}
		}
		m.quit <- true
	}()

	return m
}

// Close all metric channels so no new messages may be sent.  This is
// a prerequisite to `Wait`ing.
func (m *CollatedMetrics) Close() {
	for _, ch := range m.counters {
		close(ch)
	}
	for _, ch := range m.gauges {
		close(ch)
	}
	for _, ch := range m.customCounters {
		close(ch)
	}
	for _, ch := range m.customGauges {
		close(ch)
	}
}

// Get (possibly by creating) a counter channel by the given name.
func (m *CollatedMetrics) GetCounter(name string) chan int64 {
	ch, ok := m.counters[name]
	if ok {
		return ch
	}
	return m.NewCounter(name)
}

// Get (possibly by creating) a custom counter channel by the given name.
func (m *CollatedMetrics) GetCustomCounter(name string) chan map[string]int64 {
	ch, ok := m.customCounters[name]
	if ok {
		return ch
	}
	return m.NewCustomCounter(name)
}

// Get (possibly by creating) a custom gauge channel by the given name.
func (m *CollatedMetrics) GetCustomGauge(name string) chan map[string]int64 {
	ch, ok := m.customGauges[name]
	if ok {
		return ch
	}
	return m.NewCustomGauge(name)
}

// Get (possibly by creating) a gauge channel by the given name.
func (m *CollatedMetrics) GetGauge(name string) chan int64 {
	ch, ok := m.gauges[name]
	if ok {
		return ch
	}
	return m.NewGauge(name)
}

// Create a counter channel by the given name.
func (m *CollatedMetrics) NewCounter(name string) chan int64 {
	ch := make(chan int64)
	m.counters[name] = ch
	go m.newMetric(name, ch, m.collateCounters)
	return ch
}

// Create a custom counter channel by the given name.
func (m *CollatedMetrics) NewCustomCounter(name string) chan map[string]int64 {
	ch := make(chan map[string]int64)
	m.customCounters[name] = ch
	go m.newMetric(name, ch, m.collateCounters)
	return ch
}

// Create a custom gauge channel by the given name.
func (m *CollatedMetrics) NewCustomGauge(name string) chan map[string]int64 {
	ch := make(chan map[string]int64)
	m.customGauges[name] = ch
	go m.newMetric(name, ch, m.collateGauges)
	return ch
}

// Create a gauge channel by the given name.
func (m *CollatedMetrics) NewGauge(name string) chan int64 {
	ch := make(chan int64)
	m.gauges[name] = ch
	go m.newMetric(name, ch, m.collateGauges)
	return ch
}

// Wait for all outstanding HTTP requests to finish.  This must be called
// after `Close` has been called.
func (m *CollatedMetrics) Wait() {
	<-m.quit
}

// Serialize an `application/json` request body and do one HTTP roundtrip
// using the `http` package's `DefaultClient`.  This wrapper supplies the
// appropriate Librato Metrics API endpoint, sets the `Content-Type` header
// to `application/json`, and sets the `Authorization` header for  HTTP Basic
// authentication from the `CollatedMetrics` struct.
func (m *CollatedMetrics) do(body map[string]interface{}) error {
	b, err := json.Marshal(body)
	if nil != err {
		return err
	}
	req, err := http.NewRequest(
		"POST",
		"https://metrics-api.librato.com/v1/metrics",
		bytes.NewBuffer(b),
	)
	if nil != err {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", uaString)
	req.SetBasicAuth(m.user, m.token)
	_, err = http.DefaultClient.Do(req)
	return err
}

// Create a metric channel and begin processing messages sent
// to it in a background goroutine.
func (m *CollatedMetrics) newMetric(name string,
	i interface{},
	ch chan map[string]interface{}) {

	m.running <- true
	for {
		body := map[string]interface{}{"name": name}
		if "" != m.source {
			body["source"] = m.source
		}

		if !handle(i, body) {
			break
		}
		if _, present := body["measure_time"]; !present {
			body["measure_time"] = time.Now().Unix()
		}
		ch <- body
	}
	m.running <- false
}
