package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
	"ws-product-golang/src/server/counters"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
)

var (
	c = counters.Counter{}

	cM = make(map[string]counters.Counter)

	content = []string{"sports", "entertainment", "business", "education"}
)

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to EQ Works 😎")
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	data := content[rand.Intn(len(content))]

	key := fmt.Sprintf("%v:%v", data, time.Now().Format("2006-01-02 15:04"))
	counter, ok := cM[key]
	if ok {
		c = counter
	}

	c.Lock()
	c.View++
	c.Unlock()

	err := processRequest(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
		return
	}

	// simulate random click call
	if rand.Intn(100) < 50 {
		processClick(data)
	}

	cM[key] = c

	//reset counter
	c.Lock()
	c.View = 0
	c.Click = 0
	c.Unlock()
}

func processRequest(r *http.Request) error {
	time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)
	return nil
}

func processClick(data string) error {
	c.Lock()
	c.Click++
	c.Unlock()

	return nil
}

func statsHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	resultCh := make(chan []*counters.CounterFull)
	errCh := make(chan error)

	go func() {

		//check if context cancelled before time to talk to store
		if ctx.Err() != nil {
			return
		}

		counterList, err := counters.GetCounterStore()

		//check if context cancelled by the time done talking to DB
		if ctx.Err() != nil {
			return
		}

		if err != nil {
			errCh <- err
			return
		}

		resultCh <- counterList
		return

	}()

	select {
	case <-ctx.Done():
		log.Println(ctx.Err())
		http.Error(w, "We could not process your request at this time. Please try again later.", http.StatusRequestTimeout)
		return
	case err := <-errCh:
		log.Println(err)
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	case stats := <-resultCh:
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(stats)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}
		return
	}

}

func uploadCounters(tick int) error {
	for {
		// we create a new ticker that ticks according to update time
		ticker := time.NewTicker(time.Duration(tick) * time.Second)

		// every time our ticker ticks
		for range ticker.C {

			err := counters.AddCounter(cM)
			if err != nil {
				log.Panicln("error adding to store")
			}
			//reset map
			for k := range cM {
				delete(cM, k)
			}
		}
	}
}

func main() {

	go uploadCounters(5)

	http.HandleFunc("/", welcomeHandler)
	http.HandleFunc("/view/", viewHandler)

	//call new limimter
	lmt := newLimiter()
	http.Handle("/stats/", tollbooth.LimitFuncHandler(lmt, statsHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

//newLimiter sets up tollbooth rate limiter
func newLimiter() *limiter.Limiter {

	lmt := tollbooth.NewLimiter(2, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})

	lmt.SetIPLookups([]string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"})
	lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
		log.Println("request limit reached")
	})

	return lmt

}
