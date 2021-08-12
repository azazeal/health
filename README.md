[![Build Status](https://github.com/azazeal/health/actions/workflows/build.yml/badge.svg)](https://github.com/azazeal/health/actions/workflows/build.yml)
[![Coverage Report](https://coveralls.io/repos/github/azazeal/health/badge.svg?branch=master)](https://coveralls.io/github/azazeal/health?branch=master)
[![Go Reference](https://pkg.go.dev/badge/github.com/health/fly.svg)](https://pkg.go.dev/github.com/azazeal/health)

# health

Package health implements a reusable, concurrent health check for Go apps.

## Usage

```go
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/azazeal/health"
	"github.com/azazeal/pause"
)

func main() {
	var hc health.Check
	go longRunningTaskThatMightFail(context.TODO(), &hc)

	http.Handle("/health", &hc) // /health returns 204 if hc.Healthy(), or 503
	if err := http.ListenAndServe(":8080", nil); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func longRunningTaskThatMightFail(ctx context.Context, hc *health.Check) {
	for ctx.Err() == nil {
		if err := mightFail(ctx); err != nil {
			// the component failed, unset it and retry
			hc.Unset("component") 

			pause.For(ctx, time.Second)

			continue
		}
		hc.Set("component") // the component did not fail; carry on

		// ...
	}
}

func mightFail(context.Context) (err error) {
	// ...
	return
}
```
