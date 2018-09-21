package conduit

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	log "github.com/sirupsen/logrus"
)

// ManiphestResponse a Maniphest response
type ManiphestResponse struct {
	ID                 string      `json:"id"`
	Phid               string      `json:"phid"`
	AuthorPHID         string      `json:"authorPHID"`
	OwnerPHID          string      `json:"ownerPHID"`
	CcPHIDs            []string    `json:"ccPHIDs"`
	Status             string      `json:"status"`
	StatusName         string      `json:"statusName"`
	IsClosed           bool        `json:"isClosed"`
	Priority           string      `json:"priority"`
	PriorityColor      string      `json:"priorityColor"`
	Title              string      `json:"title"`
	Description        string      `json:"description"`
	ProjectPHIDs       []string    `json:"projectPHIDs"`
	URI                string      `json:"uri"`
	Auxiliary          interface{} `json:"auxiliary"`
	ObjectName         string      `json:"objectName"`
	DateCreated        string      `json:"dateCreated"`
	DateModified       string      `json:"dateModified"`
	DependsOnTaskPHIDs []string    `json:"dependsOnTaskPHIDs"`
}

// ManiphestQueryResponse a Query response
type ManiphestQueryResponse struct {
	Result    map[string]ManiphestResponse `json:"result"`
	ErrorCode string                       `json:"error_code"`
	ErrorInfo string                       `json:"error_info"`
}

// ManiphestTransactionResponse a Maniphest Transaction response
type ManiphestTransactionResponse struct {
	Result    map[string][]TransactionResponse `json:"result"`
	ErrorCode string                           `json:"error_code"`
	ErrorInfo string                           `json:"error_info"`
}

// TransactionResponse a Transaction response
type TransactionResponse struct {
	TaskID          string `json:"taskID"`
	TransactionID   string `json:"transactionID"`
	TransactionPHID string `json:"transactionPHID"`
	TransactionType string `json:"transactionType"`
	OldValue        string `json:"oldValue"`
	NewValue        string `json:"newValue"`
	Comments        string `json:"comments"`
	AuthorPHID      string `json:"authorPHID"`
	DateCreated     string `json:"dateCreated"`
}

// QueryManiphest maniphests from a list of PHID
func (conduit *Conduit) QueryManiphest(phids []string) (resp []ManiphestResponse, err error) {
	body := url.Values{}
	for i, phid := range phids {
		body.Add(fmt.Sprintf("phids[%d]", i), phid)
	}
	log.Debugf("maniphest.query -d %v", body)
	res, err := conduit.post("maniphest.query", body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var resBody ManiphestQueryResponse
	json.NewDecoder(res.Body).Decode(&resBody)
	if resBody.ErrorInfo != "" {
		return nil, fmt.Errorf("%s %s", resBody.ErrorCode, resBody.ErrorInfo)
	}
	if len(resBody.Result) == 0 {
		return nil, errors.New("Empty result")
	}
	values := make([]ManiphestResponse, 0, len(resBody.Result))
	for _, val := range resBody.Result {
		values = append(values, val)
	}
	return values, nil
}

// GetTransactions maniphests task transaction from a task id
func (conduit *Conduit) GetTransactions(taskIds []string) (resp map[string][]TransactionResponse, err error) {
	body := url.Values{}
	for i, phid := range taskIds {
		body.Add(fmt.Sprintf("ids[%d]", i), phid)
	}
	log.Debugf("maniphest.gettasktransactions -d %v", body)
	res, err := conduit.post("maniphest.gettasktransactions", body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// dump, e := httputil.DumpResponse(res, true)
	// if e != nil {
	// 	fmt.Println(e)
	// }
	// fmt.Println(string(dump))

	var resBody ManiphestTransactionResponse
	json.NewDecoder(res.Body).Decode(&resBody)
	if resBody.ErrorInfo != "" {
		return nil, fmt.Errorf("%s %s", resBody.ErrorCode, resBody.ErrorInfo)
	}
	if len(resBody.Result) == 0 {
		return nil, errors.New("Empty result")
	}
	return resBody.Result, nil
}
