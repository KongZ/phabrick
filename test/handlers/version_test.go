package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersion(t *testing.T) {
	w := httptest.NewRecorder()
	release := "0.1.0"
	h := version(release)
	h(w, nil)

	resp := w.Result()
	if have, want := resp.StatusCode, http.StatusOK; have != want {
		t.Errorf("Status code is wrong. Have: %d, want: %d.", have, want)
	}

	greeting, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	info := struct {
		Release string `json:"release"`
	}{}
	err = json.Unmarshal(greeting, &info)
	if err != nil {
		t.Fatal(err)
	}
	if info.Release != release {
		t.Errorf("Release version is wrong. Have: %s, want: %s", info.Release, release)
	}
}
