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

	res, err := api.doRequest(method, endpoint, payload)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// TODO: Check content head for text/html

	logrus.Debug("ResponseBody:\n", IndentJson(body))

	if res.StatusCode > 399 {
		var resErr responseError
		err = json.Unmarshal(body, &resErr)
		if err != nil {
			return nil, err
		}
		logResponse(api.Host+endpoint, res.StatusCode, resErr)

		validation := NewValidation(resErr.Message + "\n Host: " + api.Host)
		for _, re := range resErr.Items {
			validation.FormatValidationError(&re)
		}
		return validation, errors.New("Error from server")
	}

	data, err := unmarshal(body)
	if err != nil {
		return nil, err
	}
	logResponse(api.Host+endpoint, res.StatusCode, data)

	return nil, nil
}

func logResponse(url string, status int, res ResponseBody) {
	logrus.WithFields(logrus.Fields{
		"status":  status,
		"url":     url,
		"success": res.GetSuccess(),
		"message": res.GetMessage(),
		"count":   res.GetCount(),
	}).Info("Response")
}

func (api *BooberClient) doRequest(method string, endpoint string, payload []byte) (*http.Response, error) {

	url := api.Host + endpoint
	reqLog := logrus.WithFields(logrus.Fields{
		"method": method,
		"url":    url,
	})

	reqLog.Info("Request")
	if len(payload) == 0 {
		reqLog.Debug("No payload")
	} else {
		logrus.Debug("Payload:\n", IndentJson(payload))
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+api.Token)

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Error connecting to Boober")
	}

	return res, nil
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
