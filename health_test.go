package health

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strconv"
	"testing"
)

func TestFromContextPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	_ = FromContext(context.Background())
}

func TestFromContext(t *testing.T) {
	exp := new(Check)

	ctx := NewContext(context.Background(), exp)
	if got := FromContext(ctx); exp != got {
		t.Fatalf("\nexp: %p %#v\ngot: %p %#v", exp, exp, got, got)
	}
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

	empty := new(Check)
	empty.Pass()
	empty.Fail("1")
	empty.Fail("2")
	empty.Fail("3", "4", "5")
	empty.Pass("3", "4", "5")
	empty.Pass("2")
	empty.Pass("1")
	empty.Fail()

	healthy := new(Check)
	healthy.Pass("1", "3")
	healthy.Fail("2")
	healthy.Pass("2")

	unhealthy := new(Check)
	unhealthy.Fail("1", "2", "3")
	unhealthy.Pass("2")

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
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("failed reading response body: %v", err)
			}
			if kase.code != res.StatusCode {
				t.Errorf("code mismatch:\nexp: %d\ngot: %d", kase.code, res.StatusCode)
			}
			if got := string(body); kase.body != got {
				t.Errorf("body mismatch:\nexp: %q\ngot: %q", kase.body, got)
			}
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
	exp := []string{"1", "2", "3"}

	var c Check
	c.Fail(exp...)

	got := c.Failing(nil)
	sort.Strings(got)

	if !reflect.DeepEqual(exp, got) {
		t.Errorf("slice mismatch:\nexp: %v\ngot: %v", exp, got)
	}
}
