package counters

var counterStore []*CounterFull

//GetCounterStore gets from our store
func GetCounterStore() ([]*CounterFull, error) {

	return counterStore, nil

}

//AddCounter adds CounterFull to counterStore
func AddCounter(cM map[string]Counter) error {

	for k, v := range cM {
		cF := CounterFull{
			Key: k,
			C:   v,
		}
		counterStore = append(counterStore, &cF)
	}

	return nil

}
