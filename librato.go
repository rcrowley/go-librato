// Go client for Librato Metrics
//
// <https://github.com/rcrowley/go-librato>
package librato

type Metrics interface {
	Close()
	GetCounter(name string) chan int64
	GetCustomCounter(name string) chan map[string]int64
	GetCustomGauge(name string) chan map[string]int64
	GetGauge(name string) chan int64
	NewCounter(name string) chan int64
	NewCustomCounter(name string) chan map[string]int64
	NewCustomGauge(name string) chan map[string]int64
	NewGauge(name string) chan int64
	Wait()
}

func handle(i interface{}, body map[string]interface{}) bool {
	var obj map[string]int64
	var ok bool
	switch ch := i.(type) {
	case chan int64:
		body["value"], ok = <-ch
	case chan map[string]int64:
		obj, ok = <-ch
		for k, v := range obj { body[k] = v }
	}
	return ok
}
