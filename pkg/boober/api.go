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

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Count   int    `json:"count"`
}

func (r Response) GetSuccess() bool {
	return r.Success
}

func (r Response) GetMessage() string {
	return r.Message
}

func (r Response) GetCount() int {
	return r.Count
}

type ResponseBody interface {
	GetSuccess() bool
	GetMessage() string
	GetCount() int
}

type Api struct {
	Host        string
	Token       string
	Affiliation string
}

func NewApi(host, token, affiliation string) *Api {
	return &Api{
		Host:        host,
		Token:       token,
		Affiliation: affiliation,
	}
}

type HandleResponseFunc func([]byte) (ResponseBody, error)

func (api *Api) WithRequest(method string, endpoint string, payload []byte, handle HandleResponseFunc) error {

	body, status, err := api.performRequest(method, endpoint, payload)
	if err != nil {
		return err
	}

	res, err := handle(body)

	logrus.WithFields(logrus.Fields{
		"status":  status,
		"success": res.GetSuccess(),
		"message": res.GetMessage(),
		"count":   res.GetCount(),
	}).Info("Response")

	logrus.Debug("ResponseBody:\n", string(body))

	return err
}

func (api *Api) performRequest(method string, endpoint string, payload []byte) ([]byte, int, error) {

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
		return []byte{}, -1, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+api.Token)

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return []byte{}, -1, errors.Wrap(err, "Error connecting to Boober")
	}

	if res.StatusCode > 399 {
		return []byte{}, res.StatusCode, errors.New("Boober request returned with error code " + res.Status)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	return body, res.StatusCode, err
}

func IndentJson(data []byte) string {
	var out bytes.Buffer
	err := json.Indent(&out, data, "", "  ")
	if err != nil {
		logrus.Warn("Failed to indent json ", err.Error())
		return string(data)
	}

	return string(out.Bytes())
}
