package client

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

type ApiClient struct {
	Host        string
	Token       string
	Affiliation string
}

func NewApiClient(host, token, affiliation string) *ApiClient {
	return &ApiClient{
		Host:        host,
		Token:       token,
		Affiliation: affiliation,
	}
}

type UnmarshalResponseFunc func(body []byte) (ResponseBody, error)

func (api *ApiClient) Call(method string, endpoint string, payload []byte, unmarshal UnmarshalResponseFunc) (*ErrorResponse, error) {

	res, err := api.doRequest(method, endpoint, payload)
	if err != nil {
		return NewErrorResponse("Request failed"), err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return NewErrorResponse("ErrorResponse format error"), err
	}

	// TODO: Check content head for text/html

	logrus.Debug("ResponseBody", string(body))

	if res.StatusCode > 399 {
		var resErr responseError
		err = json.Unmarshal(body, &resErr)
		if err != nil {
			return nil, err
		}
		logResponse("ErrorResponse", api.Host+endpoint, res.StatusCode, resErr)

		errorResponse := NewErrorResponse(resErr.Message + "\n Host: " + api.Host)
		for _, re := range resErr.Items {
			errorResponse.FormatValidationError(&re)
		}
		return errorResponse, errors.New("Error from server")
	}

	data, err := unmarshal(body)
	if err != nil {
		return nil, err
	}
	logResponse("Response", api.Host+endpoint, res.StatusCode, data)

	return nil, nil
}

func (api *ApiClient) doRequest(method string, endpoint string, payload []byte) (*http.Response, error) {

	url := api.Host + endpoint
	logrus.WithFields(logrus.Fields{
		"method": method,
		"url":    url,
	}).Info("Request")

	if len(payload) == 0 {
		logrus.Debug("No payload")
	} else {
		logrus.Debug("Payload", string(payload))
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

func logResponse(message, url string, status int, res ResponseBody) {
	logrus.WithFields(logrus.Fields{
		"status":  status,
		"url":     url,
		"success": res.GetSuccess(),
		"message": res.GetMessage(),
		"count":   res.GetCount(),
	}).Info(message)
}
