package testhelper

import (
	"net/http"
	"net/http/httptest"
	"strings"
)

func DoHttpRequest(handler http.Handler, method string, urlStr string, json string) (w *httptest.ResponseRecorder, err error) {
	body := strings.NewReader(json)
	req, err := http.NewRequest(method, urlStr, body)
	req.Header.Set("User-Agent", "test")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return
}

func DoHttpRequestWithHeaders(handler http.Handler, method string, headers map[string]string, urlStr string, json string) (w *httptest.ResponseRecorder, err error) {
	body := strings.NewReader(json)
	req, err := http.NewRequest(method, urlStr, body)
	req.Header.Set("User-Agent", "test")
	for key, val := range headers {
		req.Header.Set(key, val)
	}
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return
}

func DoHttpRequestWithCookie(handler http.Handler, method string, cookie *http.Cookie, urlStr string, json string) (w *httptest.ResponseRecorder, err error) {
	body := strings.NewReader(json)
	req, err := http.NewRequest(method, urlStr, body)
	req.Header.Set("User-Agent", "test")
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return
}

func DoHttpAsyncRequest(handler http.Handler, method string, urlStr string, json string) (w *httptest.ResponseRecorder, err error) {
	body := strings.NewReader(json)
	req, err := http.NewRequest(method, urlStr, body)
	req.Header.Set("User-Agent", "test")
	w = httptest.NewRecorder()
	go handler.ServeHTTP(w, req)
	return
}
