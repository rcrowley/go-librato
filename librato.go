// Go client for Librato Metrics
//
// <https://github.com/rcrowley/go-librato>
package librato

// TODO
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
