package client

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type (
	// BooberResponse holds the response of an external call to boober
	BooberResponse struct {
		Success bool            `json:"success"`
		Message string          `json:"message"`
		Items   json.RawMessage `json:"items"`
		Count   int             `json:"count"`
	}

	// ErrorResponse is a structured error response
	ErrorResponse struct {
		message            string
		ContainsError      bool
		GenericErrors      []string
		IllegalFieldErrors []string
		MissingFieldErrors []string
		InvalidFieldErrors []string
	}

	errorField struct {
		Path     string      `json:"path"`
		FileName string      `json:"fileName"`
		Value    interface{} `json:"value"`
	}

	responseErrorItem struct {
		Application string `json:"application"`
		Environment string `json:"environment"`
		Details     []struct {
			Type    string     `json:"type"`
			Message string     `json:"message"`
			Field   errorField `json:"field"`
		} `json:"details"`
	}
)

func (e errorField) getValue() string {
	if e.Value != nil {
		return fmt.Sprintf("%v", e.Value)
	}
	return ""
}

// ParseItems unmarshals a boober response if it was successful
func (res *BooberResponse) ParseItems(data interface{}) error {
	if !res.Success {
		return errors.New(res.Message)
	}

	return json.Unmarshal(res.Items, data)
}

// ParseFirstItem unmarshals the first item of a boober response
func (res *BooberResponse) ParseFirstItem(data interface{}) error {
	var items []json.RawMessage
	err := res.ParseItems(&items)
	if err != nil {
		return err
	}

	if len(items) < 1 {
		return errors.New("no items available")
	}

	return json.Unmarshal(items[0], data)
}

// Error returns the error of a BooberResponse if there was an error
func (res *BooberResponse) Error() error {
	errRes, err := res.toErrorResponse()
	if err != nil {
		return err
	}
	if errRes != nil {
		return errors.New(errRes.String())
	}
	return nil
}

// String returns the ErrorResponse as a string
func (e *ErrorResponse) String() string {
	var status string

	if e.ContainsError {
		status += fmt.Sprintf("%s\n", e.message)
	}

	messages := e.getAllErrors()
	for i, message := range messages {
		status += message
		if i != len(messages)-1 {
			status += "\n\n"
		}
	}

	return status
}

func (res *BooberResponse) toWarningResponse() ([]string, error) {
	var rei []responseErrorItem
	err := json.Unmarshal(res.Items, &rei)
	if err != nil {
		return nil, err
	}

	warningFormat := `Application: %s/%s
Warning:     %s`

	var warnings []string
	for _, res := range rei {
		for _, details := range res.Details {
			warning := fmt.Sprintf(warningFormat, res.Environment, res.Application, details.Message)
			warnings = append(warnings, warning)
		}
	}

	return warnings, nil
}

func (res *BooberResponse) toErrorResponse() (*ErrorResponse, error) {
	var rei []responseErrorItem
	err := json.Unmarshal(res.Items, &rei)
	if err != nil {
		return nil, err
	}

	errorResponse := &ErrorResponse{
		message:       res.Message,
		ContainsError: true,
	}

	for _, re := range rei {
		errorResponse.formatValidationError(&re)
	}

	return errorResponse, nil
}

func (e *ErrorResponse) getAllErrors() []string {
	var errorMessages []string
	errorMessages = append(errorMessages, e.GenericErrors...)
	errorMessages = append(errorMessages, e.IllegalFieldErrors...)
	errorMessages = append(errorMessages, e.InvalidFieldErrors...)
	return append(errorMessages, e.MissingFieldErrors...)
}

// TODO: Refactor this mess
func (e *ErrorResponse) formatValidationError(res *responseErrorItem) {
	illegalFieldFormat := `Application: %s/%s
Filename:    %s
Field:       %s (Illegal)
Value:       %s
Message:     %s`

	missingFieldFormat := `Application: %s/%s
Field:       %s (Missing)
Message:     %s`

	invalidFieldFormat := `Application: %s/%s
Filename:    %s
Field:       %s (Invalid)
Message:     %s`

	genericFormat := `Application: %s/%s
Message:     %s`

	for _, message := range res.Details {
		switch message.Type {

		case "ILLEGAL":
			{
				illegal := fmt.Sprintf(illegalFieldFormat,
					res.Environment,
					res.Application,
					message.Field.FileName,
					message.Field.Path,
					message.Field.getValue(),
					message.Message,
				)
				e.IllegalFieldErrors = append(e.IllegalFieldErrors, illegal)
			}

		case "INVALID":
			{
				invalid := fmt.Sprintf(invalidFieldFormat,
					res.Environment,
					res.Application,
					message.Field.FileName,
					message.Field.Path,
					message.Message,
				)
				e.InvalidFieldErrors = append(e.InvalidFieldErrors, invalid)
			}

		case "MISSING":
			{
				missing := fmt.Sprintf(missingFieldFormat,
					res.Environment,
					res.Application,
					message.Field.Path,
					message.Message,
				)
				e.MissingFieldErrors = append(e.MissingFieldErrors, missing)
			}
		case "GENERIC":
			{
				generic := fmt.Sprintf(genericFormat,
					res.Environment,
					res.Application,
					message.Message,
				)
				e.GenericErrors = append(e.GenericErrors, generic)

			}
		}
	}
}
