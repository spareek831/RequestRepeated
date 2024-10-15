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

	// Function to perform the request
	makeRequest := func(threadID int, numRequest int) {
		defer wg.Done()
		client := &http.Client{}
		for {
			if *requestDuration > 0 && time.Since(start).Minutes() >= float64(*requestDuration) {
				return // Exit if the time duration is complete
			} else if *requestDuration == 0 && numRequest == 0 {
				break
			}

			tStart := time.Now()
			req, err := http.NewRequest(*httpMethod, *url, nil)
			if err != nil {
				fmt.Printf("Error forming request in thread %d: %s\n", threadID, err)
				return
			}

			req.Header.Add("Authorization", "Bearer "+*token)
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
			if *delay > 0 && (*randomizeDelay <= 0 || rand.Intn(100) > *randomizeDelay) {
				time.Sleep(time.Duration(*delay) * time.Millisecond)
			}

			if *requestDuration != 0 {
				numRequest -= 1
			}
		}
	}

	if *requestDuration > 0 {
		// Run for a specified duration using multiple threads
		for i := 0; i < *numThreads; i++ {
			wg.Add(1)
			go makeRequest(i, 0)
		}
	} else {
		// Run a fixed number of requests
		for i := 0; i < *numRequests; i++ {
			wg.Add(1)
			go makeRequest(i, 1)
		}
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
