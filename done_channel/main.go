package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type result struct {
	url    string
	exists bool
}

func checkIfExist(done <-chan struct{}, urls <-chan string) <-chan result {
	responsec := make(chan result)
	go func() {
		defer close(responsec)
		for url := range urls {
			select {
			case <-done:
				return
			default:
				res, err := http.Get(url)
				if err != nil {
					responsec <- result{url: url, exists: false}
				} else if res.StatusCode == http.StatusOK {
					responsec <- result{url: url, exists: true}
				} else {
					responsec <- result{url: url, exists: false}
				}
			}
		}
	}()

	return responsec
}

func checkIfExistWithContext(ctx context.Context, urls <-chan string) <-chan result {
	responsec := make(chan result)
	go func() {
		defer close(responsec)
		for url := range urls {
			select {
			case <-ctx.Done():
				err := ctx.Err()
				fmt.Println(err)
				return
			default:
				res, err := http.Get(url)
				if err != nil {
					responsec <- result{url: url, exists: false}
				} else if res.StatusCode == http.StatusOK {
					responsec <- result{url: url, exists: true}
				} else {
					responsec <- result{url: url, exists: false}
				}
			}
		}
	}()

	return responsec
}
func run(ctx context.Context) {
	contextTimeout, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	// buffer channel for input
	urls := make(chan string, 4)
	urls <- "https://google.com"
	urls <- "https://amazon.com"
	urls <- "https://in-valid-url.invalid"
	urls <- "https://facebook.com"
	close(urls)
	c := checkIfExistWithContext(contextTimeout, urls)
	now := time.Now()
	for result := range c {
		fmt.Printf("url: %s, exists: %v\n", result.url, result.exists)
	}
	fmt.Println(time.Since(now))
}
func main() {
	// done := make(chan struct{})
	// defer close(done)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	run(ctx)
}
