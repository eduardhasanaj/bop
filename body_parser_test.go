package bop

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestParseModel(t *testing.T) {
	type model struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		bop := New(w, r)
		var m model
		bop.ParseModel(&m)

		if m.Name != "John" || m.Age != 26 {
			t.Error("Parse model failed")
		}
	}

	payload := url.Values{}
	payload.Add("name", "John")
	payload.Add("age", "26")
	r := httptest.NewRequest("POST", "http://example.com/foo", strings.NewReader(payload.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(payload.Encode())))

	w := httptest.NewRecorder()
	handler(w, r)
}

func BenchmarkParseModel(b *testing.B) {
	type model struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		for i := 0; i < b.N; i++ {
			bop := New(w, r)
			var m model
			bop.ParseModel(&m)
		}
	}

	payload := url.Values{}
	payload.Add("name", "John")
	payload.Add("age", "26")
	r := httptest.NewRequest("POST", "http://example.com/foo", strings.NewReader(payload.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(payload.Encode())))

	w := httptest.NewRecorder()
	handler(w, r)
}

func BenchmarkManualParsing(b *testing.B) {
	type model struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		for i := 0; i < b.N; i++ {
			r.ParseForm()
			m := model{}
			m.Name = r.PostFormValue("name")
			age, _ := strconv.Atoi(r.PostFormValue("age"))
			m.Age = age
		}
	}

	payload := url.Values{}
	payload.Add("name", "John")
	payload.Add("age", "26")
	r := httptest.NewRequest("POST", "http://example.com/foo", strings.NewReader(payload.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(payload.Encode())))

	w := httptest.NewRecorder()
	handler(w, r)
}
