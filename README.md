# golang context demo

This repository is for demo how context work

## Cancellation and timeout

The Context pacakge offers a common methods to cancel requests

* explicit cancellation
* implicit cancellation based on a timeout or deadline

A context may also carry request-specific values, such as request ID

Many network or database requests, for example, take a context for cancellation

A context offers two controls:

* a channel that closes when the cancellation occurs
* an error that's readable once the channel closes

The error value tells you whether the request was cancelled or timed out

We often use the channel from Done() in select block


Contexts form an **immutable** tree structure
(goroutine-safe; changes to a context do not affect its ancestors)

Cancellation or timeout applies to the current context and its subtree

Ditto for a value

A subtree may be created with a shorter timeout

## Context as a tree structure

It's a tree of **immutable** nodes which can be extended

![image.png](https://i.imgur.com/xYf9Ogp.png)

## Context example

The Context value should always be the first parameter

```golang
// First runs a set of queries and returns the result from
// the first to response, canceling the others.
func First(ctx context.Context, urls []string) (*Result, error) {
  c := make(chanResult, len(urls))
  ctx, cancel := context.WithCancel(ctx)
  
  defer cancel() // cancel the other queries when we're done
  
  search := func(url string) {
     c <- runQuery(ctx, url)
  }
  ...
} 
```

## example

```golang
package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

type result struct {
	url     string
	err     error
	latency time.Duration
}

func get(ctx context.Context, url string, ch chan<- result) {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if resp, err := http.DefaultClient.Do(req); err != nil {
		ch <- result{url, err, 0}
	} else {
		t := time.Since(start).Round(time.Millisecond)
		ch <- result{url, nil, t}
		resp.Body.Close()
	}
}
func main() {
	results := make(chan result)
	list := []string{
		"https://amazon.com",
		"https://google.com",
		"https://nytimes.com",
		"https://wsj.com",
		"http://localhost:8080",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for _, url := range list {
		go get(ctx, url, results)
	}

	for range list {
		r := <-results

		if r.err != nil {
			log.Printf("%-20s %s\n", r.url, r.err)
		} else {
			log.Printf("%-20s %s\n", r.url, r.latency)
		}
	}
}

```

## another example

```golang
package main

import (
	"context"
	"log"
	"net/http"
	"runtime"
	"time"
)

type result struct {
	url     string
	err     error
	latency time.Duration
}

func get(ctx context.Context, url string, ch chan<- result) {
	var r result

	start := time.Now()
	ticker := time.NewTicker(1 * time.Second).C
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if resp, err := http.DefaultClient.Do(req); err != nil {
		r = result{url, err, 0}
	} else {
		t := time.Since(start).Round(time.Millisecond)
		r = result{url, nil, t}
		resp.Body.Close()
	}

	for {
		select {
		case ch <- r:
			return
		case <-ticker:
			log.Println("tick", r)
		}
	}
}

func first(ctx context.Context, urls []string) (*result, error) {
	results := make(chan result, len(urls)) // buffer to avoid leaking
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for _, url := range urls {
		go get(ctx, url, results)
	}
	select {
	case r := <-results:
		return &r, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func main() {
	list := []string{
		"https://amazon.com",
		"https://google.com",
		"https://nytimes.com",
		"https://wsj.com",
		// "http://localhost:8080",
	}
	r, _ := first(context.Background(), list)
	if r.err != nil {
		log.Printf("%-20s %s\n", r.url, r.err)
	} else {
		log.Printf("%-20s %s\n", r.url, r.latency)
	}
	time.Sleep(9 * time.Second)
	log.Println("quit anyway...", runtime.NumGoroutine(), "still running")
}

```

## Values

Context values should be data specific to a request, such as:

* a trace ID or start time (for latency calculation)
* security or authorization

**Avoid** using the context to carry "optional" parameters

Use a package-specific, private context key type(not string) to avoid collisions

## Values example

```golang
type contextKey int

const TraceKey contextKey = 1
// AddTrace is HTTP middleware to insert a trace ID into the request
func AddTrace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if traceID := r.Header.Get("X-Cloud-Trace-Context"); traceID != "" {
			ctx = context.WithValue(ctx, TraceKey, traceID)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
 	})
}
// ContextLog make a log with the trace ID as a prefix
func ContextLog(ctx context.Context, f string, args ...interface{}) {
	// reflection 
	traceID, ok := ctx.Value(TraceKey).(string)

	if ok && traceID != "" {
		f = traceID + ":" + f
	}

	log.Printf(f, args...)
}
```