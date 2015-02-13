// Go client for Librato Metrics
//
// <https://github.com/rcrowley/go-librato>
package librato

import "strings"

type Metrics interface {
	Close()
	GetCounter(name string) chan int64
	GetCustomCounter(name string) chan map[string]int64
	GetCustomGauge(name string) chan map[string]float64
	GetGauge(name string) chan float64
	NewCounter(name string) chan int64
	NewCustomCounter(name string) chan map[string]int64
	NewCustomGauge(name string) chan map[string]float64
	NewGauge(name string) chan float64
	Wait()
}

func handle(i interface{}, bodyMetric tmetric) bool {
	var obj map[string]int64
	var ok bool
	switch ch := i.(type) {
	case chan int64:
		bodyMetric["value"], ok = <-ch
	case chan map[string]int64:
		obj, ok = <-ch
		for k, v := range obj {
			bodyMetric[k] = v
		}
	}
	return ok
}

// models http://dev.librato.com/v1/post/metrics (3) Array format (JSON only)
type tbody map[string]tibody
type tibody []tmetric
type tmetric map[string]interface{}

// Check whether a string contains any spaces
func containAnySpaces(str string) bool {
	// Contains leading or trailing white space.
	if strings.TrimSpace(str) != str {
		return true
	}
	// Check if there are any spaces between words.
	if strings.Fields(str)[0] != str {
		return true
	}

	return false
}
