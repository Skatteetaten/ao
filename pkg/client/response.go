package client

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type (
	Response struct {
		Success bool            `json:"success"`
		Message string          `json:"message"`
		Items   json.RawMessage `json:"items"`
		Count   int             `json:"count"`
	}

	ErrorResponse struct {
		message            string
		ContainsError      bool
		IllegalFieldErrors []string
		MissingFieldErrors []string
		InvalidFieldErrors []string
		UniqueErrors       map[string]bool
	}

	errorField struct {
		Handler struct {
			Path string `json:"path"`
		} `json:"handler"`
		Source struct {
			Name string `json:"name"`
		} `json:"source"`
		Value interface{} `json:"value"`
	}

	responseErrorItem struct {
		Application string `json:"application"`
		Environment string `json:"environment"`
		Messages    []struct {
			Type    string     `json:"type"`
			Message string     `json:"message"`
			Field   errorField `json:"field"`
		} `json:"messages"`
	}
)

func (res *Response) ParseItems(data interface{}) error {
	if !res.Success {
		return errors.New(res.Message)
	}

	return json.Unmarshal(res.Items, data)
}

func (res *Response) ParseFirstItem(data interface{}) error {
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

func (res *Response) ToErrorResponse() (*ErrorResponse, error) {
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
		UniqueErrors:  make(map[string]bool),
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
	errorMessages := append(e.IllegalFieldErrors, e.InvalidFieldErrors...)
	return append(errorMessages, e.MissingFieldErrors...)
}

func (e *ErrorResponse) Contains(key string) bool {
	return e.UniqueErrors[key]
}

func (e *ErrorResponse) FormatValidationError(res *responseErrorItem) {
	// TODO: Structs?
	illegalFieldFormat := `Filename:    %s
Path:        %s
Value:       %s
Message:     %s`
	missingFieldFormat := `Application: %s/%s
Path:        %s (Missing)
Message:     %s`

	invalidFieldFormat := `Filename:    %s
Path:        %s
Message:     %s`

	for _, message := range res.Messages {
		k := []string{
			message.Field.Source.Name,
			message.Field.Handler.Path,
		}
		key := strings.Join(k, "|")

		if e.Contains(key) {
			continue
		}

		if message.Type != "MISSING" {
			e.UniqueErrors[key] = true
		}

		switch message.Type {
		case "ILLEGAL":
			{
				illegal := fmt.Sprintf(illegalFieldFormat,
					message.Field.Source.Name,
					message.Field.Handler.Path,
					message.Field.Value,
					message.Message,
				)
				e.IllegalFieldErrors = append(e.IllegalFieldErrors, illegal)
			}

		case "INVALID":
			{
				invalid := fmt.Sprintf(invalidFieldFormat,
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
		}
	}
}
