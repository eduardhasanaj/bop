package bop

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	jsoniter "github.com/json-iterator/go"
)

type testModel struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Active    bool   `json:"active"`
}

var (
	tModel = testModel{
		ID:        1,
		FirstName: "John",
		LastName:  "Smith",
		Username:  "smith.john2020",
		Password:  "somestrongpassword",
		Active:    true,
	}

	tModelFormValues = url.Values{
		"id":         []string{"1"},
		"first_name": []string{"John"},
		"last_name":  []string{"Smith"},
		"username":   []string{"smith.john2020"},
		"password":   []string{"somestrongpassword"},
		"active":     []string{"true"},
	}
)

func TestParseModelFromPostBody(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		bop := New(w, r)
		var m testModel
		cols, err := bop.ParseModel(&m)
		if err != nil {
			t.Errorf("Test failed with error %v", err)
		} else if len(cols) != 6 {
			t.Error("Column number is not valid")
		} else if m.FirstName != "John" || m.LastName != "Smith" || m.Username != "smith.john2020" || m.Password != "somestrongpassword" || m.Active == false {
			t.Error("Parse model failed")
		}
	}

	r := httptest.NewRequest("POST", "http://example.com/foo", strings.NewReader(tModelFormValues.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(tModelFormValues.Encode())))

	w := httptest.NewRecorder()
	handler(w, r)
}

func TestParseModelFromPostBodyFailure(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		bop := New(w, r)
		var m testModel
		_, err := bop.ParseModel(&m)
		if err == nil {
			t.Errorf("Test failed because and error should be thrown")
		}
	}

	payload := url.Values{}
	payload.Add("some random key", "some random val")
	r := httptest.NewRequest("POST", "http://example.com/foo", strings.NewReader(payload.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(payload.Encode())))

	w := httptest.NewRecorder()
	handler(w, r)
}

func TestParseModelFromJson(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		bop := New(w, r)
		var m testModel
		cols, err := bop.ParseModel(&m)
		if err != nil {
			t.Errorf("Test failed with error %v", err)
		} else if len(cols) != 6 {
			t.Error("Column number is not valid")
		} else if m.FirstName != "John" || m.LastName != "Smith" || m.Username != "smith.john2020" || m.Password != "somestrongpassword" || m.Active == false {
			t.Error("Parse model failed")
		}
	}

	json, _ := jsoniter.Marshal(&tModel)
	r := httptest.NewRequest("POST", "http://example.com/foo", bytes.NewReader(json))
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Content-Length", strconv.Itoa(len(json)))

	w := httptest.NewRecorder()
	handler(w, r)
}

func TestParseModelFromJsonFailure(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		bop := New(w, r)
		var m testModel
		_, err := bop.ParseModel(&m)
		if err == nil {
			t.Errorf("Test failed because an error should be thrown")
		}
	}

	json := []byte(`
		"key": "Value"
	`)

	r := httptest.NewRequest("POST", "http://example.com/foo", bytes.NewReader(json))
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Content-Length", strconv.Itoa(len(json)))

	w := httptest.NewRecorder()
	handler(w, r)
}

func BenchmarkParseModelPostForm(b *testing.B) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		bop := New(w, r)
		var m testModel
		_, _ = bop.ParseModel(&m)
	}

	payload := []byte(tModelFormValues.Encode())
	length := strconv.Itoa(len(tModelFormValues.Encode()))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest("POST", "http://example.com/foo", bytes.NewReader(payload))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Add("Content-Length", length)

		w := httptest.NewRecorder()
		handler(w, r)
	}
}

func BenchmarkParseModelFromJson(b *testing.B) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		bop := New(w, r)
		var m testModel
		_, _ = bop.ParseModel(&m)
	}

	json, _ := jsoniter.Marshal(&tModel)
	length := strconv.Itoa(len(json))

	b.ResetTimer()

	for j := 0; j < b.N; j++ {
		r := httptest.NewRequest("POST", "http://example.com/foo", bytes.NewReader(json))
		r.Header.Add("Content-Type", "application/json")
		r.Header.Add("Content-Length", length)

		w := httptest.NewRecorder()
		handler(w, r)
	}
}

func BenchmarkManualParseModelPostForm(b *testing.B) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyPayloadSize)
		r.ParseForm()
		var m testModel
		id, _ := strconv.Atoi(r.PostFormValue("id"))
		m.ID = id
		m.FirstName = r.PostFormValue("first_name")
		m.LastName = r.PostFormValue("last_name")
		m.Username = r.PostFormValue("username")
		m.Password = r.PostFormValue("password")
		active, _ := strconv.ParseBool(r.FormValue("active"))
		m.Active = active
	}

	payload := []byte(tModelFormValues.Encode())
	length := strconv.Itoa(len(tModelFormValues.Encode()))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest("POST", "http://example.com/foo", bytes.NewReader(payload))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Add("Content-Length", length)

		w := httptest.NewRecorder()
		handler(w, r)
	}
}

func BenchmarkManualParseModelFromJson(b *testing.B) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyPayloadSize)
		var m testModel
		d := jsoniter.NewDecoder(r.Body)
		d.Decode(&m)
	}

	json, _ := jsoniter.Marshal(&tModel)
	length := strconv.Itoa(len(json))

	b.ResetTimer()

	for j := 0; j < b.N; j++ {
		r := httptest.NewRequest("POST", "http://example.com/foo", bytes.NewReader(json))
		r.Header.Add("Content-Type", "application/json")
		r.Header.Add("Content-Length", length)

		w := httptest.NewRecorder()
		handler(w, r)
	}
}
