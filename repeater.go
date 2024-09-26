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
	numRequests := flag.Int("n", 1000, "Number of requests to repeat")
	url := flag.String("url", "", "URL to call")
	httpMethod := flag.String("method", "POST", "HTTP method, defaults to POST")
	token := flag.String("token", "", "Bearer token for authorization")
	delay := flag.Int("delay", 100, "Delay in milliseconds between each request")
	randomizeDelay := flag.Int("randomizeNoDelay", -1, "Percentage of requests to be made immediately without delay")

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

	for i := 0; i < *numRequests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			tStart := time.Now()
			client := &http.Client{}
			req, err := http.NewRequest(*httpMethod, *url, nil)
			if err != nil {
				fmt.Printf("Error in forming request %d: %s\n", i, err)
				return
			}

			req.Header.Add("Authorization", "Bearer "+*token)
			res, err := client.Do(req)
			if err != nil {
				fmt.Printf("Error from upstream in request %d: %s\n", i, err)
				return
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				fmt.Println("Error reading response:", err)
				return
			}

			durationChan <- time.Since(tStart)
			fmt.Printf("request %d body is \n %s", i, string(body))
		}(i)

		if *delay > 0 && *randomizeDelay <= 0 {
			time.Sleep(time.Duration(*delay) * time.Millisecond)
		} else if *delay > 0 && rand.Intn(100) > *randomizeDelay {
			time.Sleep(time.Duration(*delay) * time.Millisecond)
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
