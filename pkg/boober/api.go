package boober

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"bytes"
	"io/ioutil"
	"encoding/json"
	"github.com/pkg/errors"
)

type ApplicationId struct {
	Environment string `json:"environment"`
	Application string `json:"application"`
}

type BooberClient struct {
	Host        string
	Token       string
	Affiliation string
}

func NewBooberClient(host, token, affiliation string) *BooberClient {
	return &BooberClient{
		Host:        host,
		Token:       token,
		Affiliation: affiliation,
	}
}

type UnmarshalResponseFunc func(body []byte) (ResponseBody, error)

func (api *BooberClient) Call(method string, endpoint string, payload []byte, unmarshal UnmarshalResponseFunc) (*Validation, error) {

	body, status, err := api.doRequest(method, endpoint, payload)
	if err != nil {
		return nil, err
	}

	logrus.Debug("ResponseBody:\n", string(body))

	if status > 399 {
		var resErr responseError
		err = json.Unmarshal(body, &resErr)
		if err != nil {
			return nil, err
		}
		logResponse(status, resErr)

		validation := NewValidation(resErr.Message)
		for _, re := range resErr.Items {
			validation.FormatValidationError(&re)
		}
		return validation, nil
	}

	res, err := unmarshal(body)
	if err != nil {
		return nil, err
	}
	logResponse(status, res)

	return nil, nil
}

func logResponse(status int, res ResponseBody) {
	logrus.WithFields(logrus.Fields{
		"status":  status,
		"success": res.GetSuccess(),
		"message": res.GetMessage(),
		"count":   res.GetCount(),
	}).Info("Response")
}

func (api *BooberClient) doRequest(method string, endpoint string, payload []byte) ([]byte, int, error) {

	url := api.Host + endpoint
	reqLog := logrus.WithFields(logrus.Fields{
		"method": method,
		"url":    url,
	})

	reqLog.Info("Request")
	if len(payload) == 0 {
		reqLog.Debug("No payload")
	} else {
		logrus.Debug("Payload:\n", string(payload))
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		// TODO: status code as -1 for internal errors?
		return []byte{}, -1, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+api.Token)

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return []byte{}, -1, errors.Wrap(err, "Error connecting to Boober")
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	return body, res.StatusCode, err
}

// TODO: Do we need this?
func IndentJson(data []byte) string {
	var out bytes.Buffer
	err := json.Indent(&out, data, "", "  ")
	if err != nil {
		logrus.Warn("Failed to indent json ", err.Error())
		return string(data)
	}

	return string(out.Bytes())
}
