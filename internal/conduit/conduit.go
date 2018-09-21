package conduit

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/KongZ/phabrick/internal/config"
)

// Conduit new Conduit handler
type Conduit struct {
	Config *config.Config
}

// Post request to Conduit API
func (conduit *Conduit) post(method string, data url.Values) (resp *http.Response, err error) {
	tr := &http.Transport{DisableKeepAlives: true}
	client := &http.Client{Transport: tr}
	data.Set("api.token", conduit.Config.Phabricator.Token)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/%s", conduit.Config.Phabricator.URL, method), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return client.Do(req)
}
