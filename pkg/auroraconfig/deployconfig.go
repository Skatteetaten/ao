package auroraconfig

import (
	"errors"
	"strconv"

	"encoding/json"

	"github.com/skatteetaten/ao/pkg/serverapi"
)

type ApplicationResult struct {
	DeployId           string
	AuroraDc           serverapi.AuroraDeploymentConfig
	OpenShiftResponses []serverapi.OpenShiftResponse
}

func Response2ApplicationResult(response serverapi.Response) (applicationResult *ApplicationResult, err error) {
	if len(response.Items) != 1 {
		err = errors.New("Illegal response from Boober: Expected 1 Application Result record, got " + strconv.Itoa(len(response.Items)))
		return nil, err
	}

	applicationResult = new(ApplicationResult)
	err = json.Unmarshal(response.Items[0], &applicationResult)
	if err != nil {
		return nil, err
	}

	return applicationResult, nil
}
