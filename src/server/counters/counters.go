package counters

import "sync"

//Counter is our counter object
type Counter struct {
	sync.Mutex
	View  int
	Click int
}

//CounterFull is Counter including key
type CounterFull struct {
	Key string
	C   Counter
}
