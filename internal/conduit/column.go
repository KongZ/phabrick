package conduit

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	log "github.com/sirupsen/logrus"
)

// ColumnQueryResponse a Query response
type ColumnQueryResponse struct {
	Result    map[string]ColumnResponse `json:"result"`
	ErrorCode string                    `json:"error_code"`
	ErrorInfo string                    `json:"error_info"`
}

// ColumnResponse a Column response
type ColumnResponse struct {
	Phid     string `json:"phid"`
	URI      string `json:"uri"`
	TypeName string `json:"typeName"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	FullName string `json:"fullName"`
	Status   string `json:"status"`
}

// QueryColumn projects from a list of PHID
func (conduit *Conduit) QueryColumn(phids []string) (resp []ColumnResponse, err error) {
	body := url.Values{}
	for i, phid := range phids {
		body.Add(fmt.Sprintf("phids[%d]", i), phid)
	}
	log.Debugf("phid.query -d %+v", body)
	res, e := conduit.post("phid.query", body)
	if e != nil {
		return nil, e
	}

	defer res.Body.Close()
	var resBody ColumnQueryResponse
	json.NewDecoder(res.Body).Decode(&resBody)
	if resBody.ErrorInfo != "" {
		return nil, fmt.Errorf("%s %s", resBody.ErrorCode, resBody.ErrorInfo)
	}
	if len(resBody.Result) == 0 {
		return nil, errors.New("Empty result")
	}
	values := make([]ColumnResponse, 0, len(resBody.Result))
	for _, val := range resBody.Result {
		values = append(values, val)
	}
	return values, nil
}
