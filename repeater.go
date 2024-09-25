package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

func main() {
	// Define flags for the command-line parameters
	numRequests := flag.Int("n", 1000, "Number of requests to repeat")
	url := flag.String("url", "http://localhost:8090/api/1/rest/feed/run/task/snaplogic/projects/shared/%20new%20pipeline%200%20Task12", "URL to call")
	httpMethod := flag.String("method", "POST", "URL to call, defaults to POST request")
	token := flag.String("token", "", "Bearer token for authorization")
	delay := flag.Int("delay", 10, "Delay in milliseconds between each request, defaults to 1000ms")

	// Parse the flags
	flag.Parse()

	// Check if token is provided
	if *token == "" {
		fmt.Println("Please provide a valid bearer token using the -token flag.")
		return
	}

	// Set up the wait group
	wg := sync.WaitGroup{}

	for i := 0; i < *numRequests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// Create the HTTP client
			client := &http.Client{}

			// Create a new request for each iteration
			req, err := http.NewRequest(*httpMethod, *url, nil)
			if err != nil {
				fmt.Printf("err in forming request %d err is %s\n ", i, err)
				return
			}

			// Add authorization header
			req.Header.Add("Authorization", "Bearer "+*token)

			// Send the request
			res, err := client.Do(req)
			if err != nil {
				fmt.Printf("err from upstream in routing %d err is %s\n ", i, err)
				return
			}
			defer res.Body.Close()

			// Read the response body
			body, err := io.ReadAll(res.Body)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Print the response body
			fmt.Printf("request %d body is \n %s", i, string(body))
		}(i)
		// Introduce a delay between each request
		if *delay > 0 {
			time.Sleep(time.Duration(*delay) * time.Millisecond)
		}
	}

	wg.Wait() // Wait for all goroutines to finish
}
