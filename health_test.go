package health

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromContextPanics(t *testing.T) {
	assert.Panics(t, func() { _ = FromContext(context.TODO()) })
}

func TestFromContext(t *testing.T) {
	exp := new(Check)

	ctx := NewContext(context.TODO(), exp)
	assert.Same(t, exp, FromContext(ctx))
}

func TestServeHTTP(t *testing.T) {
	const (
		put  = http.MethodPut
		get  = http.MethodGet
		head = http.MethodHead

		sok = http.StatusOK
		tok = "OK\n"

		snc = http.StatusNoContent
		tnc = ""

		smna = http.StatusMethodNotAllowed
		tmna = "Method Not Allowed\n"

		ssu = http.StatusServiceUnavailable
		tsu = "Service Unavailable\n"
	)

	empty := &Check{}
	empty.Set()
	empty.Unset("1")
	empty.Unset("2")
	empty.Unset("3", "4", "5")
	empty.Set("3", "4", "5")
	empty.Set("2")
	empty.Set("1")
	empty.Unset()

	healthy := new(Check)
	healthy.Set("1", "3")
	healthy.Unset("2")
	healthy.Set("2")

	unhealthy := new(Check)
	unhealthy.Unset("1", "2", "3")
	unhealthy.Set("2")

	cases := []*testCase{
		// tests on empty
		0: newTestCase(empty, put, smna, tmna),
		1: newTestCase(empty, head, snc, tnc),
		2: newTestCase(empty, get, sok, tok),

		// tests on healthy
		3: newTestCase(healthy, head, snc, tnc),
		4: newTestCase(healthy, get, sok, tok),

		// tests on unhealthy
		5: newTestCase(unhealthy, head, ssu, tnc),
		6: newTestCase(unhealthy, get, ssu, tsu),
	}

	for caseIndex := range cases {
		kase := cases[caseIndex]

		t.Run(strconv.Itoa(caseIndex), func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(kase.method, "/state", nil)

			kase.check.ServeHTTP(rec, req)
			res := rec.Result()

			// TODO: use io.ReadAll when support for Go 1.15 is dropped
			got, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, kase.code, res.StatusCode)
			assert.Equal(t, kase.body, string(got))
		})
	}
}

func newTestCase(check *Check, method string, code int, body string) *testCase {
	return &testCase{
		check:  check,
		method: method,
		code:   code,
		body:   body,
	}
}

type testCase struct {
	check  *Check
	method string // request method
	code   int    // expected status code
	body   string // expected body
}

func TestFailing(t *testing.T) {
	dst := make([]string, 5, 8)
	for i := range dst {
		dst[i] = strconv.Itoa(i)
	}

	var c Check
	for i := len(dst); i < cap(dst); i++ {
		c.Unset(strconv.Itoa(i))
	}
	got := c.Failing(dst[len(dst):])

	p1 := reflect.ValueOf(dst[len(dst):]).Pointer()
	p2 := reflect.ValueOf(got).Pointer()
	assert.Equal(t, p1, p2)

	exp := make([]string, cap(dst))
	for i := range exp {
		exp[i] = strconv.Itoa(i)
	}
	assert.ElementsMatch(t, exp, dst[:cap(dst)])
}
