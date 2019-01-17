package conduit

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	log "github.com/sirupsen/logrus"
)

// ProjectQueryResponse a Query response
type ProjectQueryResponse struct {
	Result struct {
		Data    ProjectDataResponse `json:"data"`
		SlugMap []interface{}       `json:"slugMap"`
		Cursor  struct {
			Limit  int         `json:"limit"`
			After  interface{} `json:"after"`
			Before interface{} `json:"before"`
		} `json:"cursor"`
	} `json:"result"`
	ErrorCode string `json:"error_code"`
	ErrorInfo string `json:"error_info"`
}

// ProjectResponse a Project response
type ProjectResponse struct {
	ID               string   `json:"id"`
	Phid             string   `json:"phid"`
	Name             string   `json:"name"`
	ProfileImagePHID string   `json:"profileImagePHID"`
	Icon             string   `json:"icon"`
	Color            string   `json:"color"`
	Members          []string `json:"members"`
	Slugs            []string `json:"slugs"`
	DateCreated      string   `json:"dateCreated"`
	DateModified     string   `json:"dateModified"`
}

// ProjectDataResponse a Query response
type ProjectDataResponse map[string]ProjectResponse

// QueryProject projects from a list of PHID
func (conduit *Conduit) QueryProject(phids []string) (resp []ProjectResponse, err error) {
	body := url.Values{}
	for i, phid := range phids {
		body.Add(fmt.Sprintf("phids[%d]", i), phid)
	}
	log.Debugf("project.query -d %+v", body)
	res, e := conduit.post("project.query", body)
	if e != nil {
		return nil, e
	}

	defer res.Body.Close()
	var resBody ProjectQueryResponse
	json.NewDecoder(res.Body).Decode(&resBody)
	if resBody.ErrorInfo != "" {
		return nil, fmt.Errorf("%s %s", resBody.ErrorCode, resBody.ErrorInfo)
	}
	if len(resBody.Result.Data) == 0 {
		return nil, errors.New("Empty result")
	}
	values := make([]ProjectResponse, 0, len(resBody.Result.Data))
	for _, val := range resBody.Result.Data {
		values = append(values, val)
	}
	return values, nil
}
