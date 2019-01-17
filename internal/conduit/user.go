package conduit

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	log "github.com/sirupsen/logrus"
)

// UserQueryResponse a Query response
type UserQueryResponse struct {
	Result    []UserResponse
	ErrorCode string `json:"error_code"`
	ErrorInfo string `json:"error_info"`
}

// UserResponse an User response
type UserResponse struct {
	Phid     string   `json:"phid"`
	UserName string   `json:"userName"`
	RealName string   `json:"realName"`
	Image    string   `json:"image"`
	URI      string   `json:"uri"`
	Roles    []string `json:"roles"`
}

// QueryUser projects from a list of PHID
func (conduit *Conduit) QueryUser(phids []string) (resp []UserResponse, err error) {
	body := url.Values{}
	for i, phid := range phids {
		body.Add(fmt.Sprintf("phids[%d]", i), phid)
	}
	log.Debugf("user.query -d %+v", body)
	res, err := conduit.post("user.query", body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var resBody UserQueryResponse
	json.NewDecoder(res.Body).Decode(&resBody)
	if resBody.ErrorInfo != "" {
		return nil, fmt.Errorf("%s %s", resBody.ErrorCode, resBody.ErrorInfo)
	}
	if len(resBody.Result) == 0 {
		return nil, errors.New("Empty result")
	}
	values := make([]UserResponse, 0, len(resBody.Result))
	for _, val := range resBody.Result {
		values = append(values, val)
	}
	return values, nil
}
