package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func main() {
	start := time.Now()
	numRequests := flag.Int("n", 1, "Number of requests to repeat")
	url := flag.String("url", "", "URL to call")
	httpMethod := flag.String("method", "POST", "HTTP method, defaults to POST")
	token := flag.String("token", "", "Bearer token for authorization")
	delay := flag.Int("delay", 100, "Delay in milliseconds between each request")
	randomizeDelay := flag.Int("randomizeNoDelay", -1, "Percentage of requests to be made immediately without delay")
	requestDuration := flag.Int("requestDuration", -1, "Run requests for the given number of minutes")
	numThreads := flag.Int("numThreads", 20, "Number of threads to use when running for a specified duration")

	flag.Parse()

	if *token == "" {
		fmt.Println("Please provide a valid bearer token using the -token flag.")
		return
	}

	wg := sync.WaitGroup{}
	durationChan := make(chan time.Duration, *numRequests)
	var sum int64 = 0
	var total int = 0
	var min int64 = math.MaxInt64
	var max int64 = 0

	go func() {
		for tDuration := range durationChan {
			durationMs := tDuration.Milliseconds()
			sum += durationMs
			total++
			if durationMs < min {
				min = durationMs
			}
			if durationMs > max {
				max = durationMs
			}
		}
	}()

	if *requestDuration > 0 {
		// Run for a specified duration using multiple threads
		for i := 0; i < *numThreads; i++ {
			wg.Add(1)
			go makeRequestForDuration(i, &wg, start, *requestDuration, *url, *httpMethod, *token, *delay, *randomizeDelay, durationChan)
		}
	} else {
		// Run a fixed number of requests
		makeNRequest(*numRequests, *url, *httpMethod, *token, *delay, *randomizeDelay, durationChan, &wg)
	}

	wg.Wait()
	close(durationChan)

	duration := time.Since(start)

	var average int64 = 0
	if total > 0 {
		average = sum / int64(total)
	}

	fmt.Printf("\nFinal summary:\n")
	fmt.Printf("Total tasks: %d\n", total)
	fmt.Printf("Sum: %d ms\n", sum)
	fmt.Printf("Average per task: %d ms\n", average)
	fmt.Printf("Min per task: %d ms\n", min)
	fmt.Printf("Max per task: %d ms\n", max)
	fmt.Println("Total Duration:", duration)
	fmt.Println("Duration in seconds:", duration.Seconds())
	fmt.Println("Duration in milliseconds:", duration.Milliseconds())
}

// Function to handle repeated requests for a duration
func makeRequestForDuration(threadID int, wg *sync.WaitGroup, start time.Time,
	requestDuration int, url, httpMethod, token string, delay, randomizeDelay int, durationChan chan time.Duration) {
	defer wg.Done()
	client := &http.Client{}

	for {
		// Check if the duration has been reached
		if time.Since(start).Minutes() >= float64(requestDuration) {
			break
		}
		makeHTTPRequest(threadID, client, url, httpMethod, token, delay, randomizeDelay, durationChan)
	}
}

// Function to handle N requests with multiple threads
func makeNRequest(numRequests int, url, httpMethod, token string, delay, randomizeDelay int, durationChan chan time.Duration, wg *sync.WaitGroup) {
	client := &http.Client{}
	numThreads := numRequests / 3

	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()
			for j := 0; j < numRequests/numThreads; j++ {
				makeHTTPRequest(threadID, client, url, httpMethod, token, delay, randomizeDelay, durationChan)
			}
		}(i)
	}
}

// Function to perform the HTTP request
func makeHTTPRequest(threadID int, client *http.Client, url, httpMethod, token string, delay, randomizeDelay int, durationChan chan time.Duration) {
	tStart := time.Now()
	req, err := http.NewRequest(httpMethod, url, nil)
	if err != nil {
		fmt.Printf("Error forming request in thread %d: %s\n", threadID, err)
		return
	}

	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error from upstream in thread %d: %s\n", threadID, err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	durationChan <- time.Since(tStart)
	fmt.Printf("Thread %d response:\n%s\n", threadID, string(body))

	// Handle delay and randomization
	if delay > 0 && (randomizeDelay <= 0 || rand.Intn(100) > randomizeDelay) {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}
