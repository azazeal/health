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

	"github.com/azazeal/health"
)

func main() {
	var hc health.Checker
	go longRunningTaskThatMightFail(context.TODO(), &hc)

	http.Handle("/health", &hc) // /health returns 204 if hc.Healthy(), or 503
	if err := http.ListenAndServe(":8080", nil); err != http.ErrServerClosed {
		log.Fatal(err)
	}
	hc.Healthy()
}

func longRunningTaskThatMightFail(ctx context.Context, hc *health.Checker) {
	for ctx.Err() == nil {
		if err := mightFail(ctx); err != nil {
			hc.Unset("component") // the component failed, unset it.
			continue
		}
		hc.Set("component") // the component did not fail.

		// continue on
	}

	hc.Healthy()
}

func mightFail(context.Context) (err error) {
	// do something that might fail
	// ...
	return
}
```
