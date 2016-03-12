package session

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockReader struct {
}

func (r *mockReader) Read(p []byte) (n int, err error) {
	return len(p), nil
}

func TestGetSessionID(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "/", &mockReader{})

	if err != nil {
		t.Errorf("error %v\n", err)
	}

	sessionID := getSessionID(r)

	if sessionID == "" {
		t.Errorf("expected string exists, actual %v", sessionID)
	}

	sessionID2 := getSessionID(r)

	if sessionID == sessionID2 {
		t.Errorf("expected always differ sessionID")
	}
}

type mockHandler struct {
	session []byte
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", h.session)
}

func (h *mockHandler) GetSession(session []byte) {
	h.session = session
}

func (h *mockHandler) SetSession() []byte {
	return []byte("foo")
}

func TestHandler(t *testing.T) {
	Setup("exampleCookieKey", "exampleNamespace", MemoryDialect, nil)
	defer Dispose()

	ts := httptest.NewServer(Sugar(&mockHandler{}))

	// first request: nothing stores
	r, err := http.Get(ts.URL)
	if err != nil {
		panic(err)
	}

	data, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	if string(data) != "" {
		t.Errorf("Response error: expected empty, actual %v", string(data))
	}

	client := &http.Client{}
	cookie := r.Cookies()[0]
	request, err := http.NewRequest(http.MethodGet, ts.URL, &mockReader{})
	if err != nil {
		panic(err)
	}
	request.Header.Set("Cookie", cookie.String())
	// second request: expect to response the session when first requests
	r, err = client.Do(request)
	if err != nil {
		panic(err)
	}

	data, err = ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	if "foo" != string(data) {
		t.Errorf("Response error: expected %v, actual %v", "foo", string(data))
	}

}
