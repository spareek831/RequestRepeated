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
	numRequests := flag.Int("n", 1000, "Number of requests to repeat")
	// You can add the default over here if needed
	url := flag.String("url", "", "URL to call")
	httpMethod := flag.String("method", "POST", "URL to call, defaults to POST request")
	token := flag.String("token", "", "Bearer token for authorization")
	delay := flag.Int("delay", 10, "Delay in milliseconds between each request, defaults to 1000ms")

	flag.Parse()

	if *token == "" {
		fmt.Println("Please provide a valid bearer token using the -token flag.")
		return
	}

	wg := sync.WaitGroup{}

	for i := 0; i < *numRequests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			
			client := &http.Client{}
			req, err := http.NewRequest(*httpMethod, *url, nil)
			if err != nil {
				fmt.Printf("err in forming request %d err is %s\n ", i, err)
				return
			}
			
			req.Header.Add("Authorization", "Bearer "+*token)
			res, err := client.Do(req)
			if err != nil {
				fmt.Printf("err from upstream in routing %d err is %s\n ", i, err)
				return
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				fmt.Println(err)
				return
			}
			
			fmt.Printf("request %d body is \n %s", i, string(body))
		}(i)
		if *delay > 0 {
			time.Sleep(time.Duration(*delay) * time.Millisecond)
		}
	}

	wg.Wait()
}
