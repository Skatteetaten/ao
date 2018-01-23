package client

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type (
	BooberResponse struct {
		Success bool            `json:"success"`
		Message string          `json:"message"`
		Items   json.RawMessage `json:"items"`
		Count   int             `json:"count"`
	}

	ErrorResponse struct {
		message            string
		ContainsError      bool
		GenericErrors      []string
		IllegalFieldErrors []string
		MissingFieldErrors []string
		InvalidFieldErrors []string
	}

	errorField struct {
		Handler struct {
			Path string `json:"path"`
		} `json:"handler"`
		Source struct {
			Name string `json:"name"`
		} `json:"source"`
		Value          interface{} `json:"value"`
		DefaultOrValue interface{} `json:"defaultOrValue"`
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

func (e errorField) GetValue() string {
	if e.Value != nil {
		return fmt.Sprintf("%v", e.Value)
	}
	if e.DefaultOrValue != nil {
		return fmt.Sprintf("%v", e.DefaultOrValue)
	}
	return ""
}

func (res *BooberResponse) ParseItems(data interface{}) error {
	if !res.Success {
		return errors.New(res.Message)
	}

	return json.Unmarshal(res.Items, data)
}

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

func (res *BooberResponse) ToErrorResponse() (*ErrorResponse, error) {
	var rei []responseErrorItem
	err := json.Unmarshal(res.Items, &rei)
	if err != nil {
		return nil, err
	}

	errorResponse := NewErrorResponse(res.Message)
	for _, re := range rei {
		errorResponse.FormatValidationError(&re)
	}

	return errorResponse, nil
}

func NewErrorResponse(message string) *ErrorResponse {
	return &ErrorResponse{
		message:       message,
		ContainsError: true,
	}
}

func (e *ErrorResponse) String() string {
	var status string

	if e.ContainsError {
		status += fmt.Sprintf("%s\n", e.message)
	}

	messages := e.GetAllErrors()
	for i, message := range messages {
		status += message
		if i != len(messages)-1 {
			status += "\n\n"
		}
	}

	return status
}

func (e *ErrorResponse) SetMessage(m string) {
	e.message = m
	e.ContainsError = true
}

func (e *ErrorResponse) GetAllErrors() []string {
	var errorMessages []string
	errorMessages = append(errorMessages, e.GenericErrors...)
	errorMessages = append(errorMessages, e.IllegalFieldErrors...)
	errorMessages = append(errorMessages, e.InvalidFieldErrors...)
	return append(errorMessages, e.MissingFieldErrors...)
}

// TODO: Refactor this mess
func (e *ErrorResponse) FormatValidationError(res *responseErrorItem) {
	illegalFieldFormat := `Application: %s/%s
Filename:    %s
Field:       %s
Value:       %s
Message:     %s`

	missingFieldFormat := `Application: %s/%s
Field:       %s (Missing)
Message:     %s`

	invalidFieldFormat := `
Application: %s/%s
Filename:    %s
Field:       %s
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
					message.Field.Source.Name,
					message.Field.Handler.Path,
					message.Field.GetValue(),
					message.Message,
				)
				e.IllegalFieldErrors = append(e.IllegalFieldErrors, illegal)
			}

		case "INVALID":
			{
				invalid := fmt.Sprintf(invalidFieldFormat,
					res.Environment,
					res.Application,
					message.Field.Source.Name,
					message.Field.Handler.Path,
					message.Message,
				)
				e.InvalidFieldErrors = append(e.InvalidFieldErrors, invalid)
			}

		case "MISSING":
			{
				missing := fmt.Sprintf(missingFieldFormat,
					res.Environment,
					res.Application,
					message.Field.Handler.Path,
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
