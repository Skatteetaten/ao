package client

import (
	"fmt"
	"strings"
)

type responseErrorItem struct {
	Application string `json:"application"`
	Environment string `json:"environment"`
	Messages    []struct {
		Type    string `json:"type"`
		Message string `json:"message"`
		Field   struct {
			Path   string `json:"path"`
			Value  string `json:"value"`
			Source string `json:"source"`
		} `json:"field"`
	} `json:"messages"`
}

type ResponseBody interface {
	GetSuccess() bool
	GetMessage() string
	GetCount() int
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

type ErrorResponse struct {
	message            string
	ContainsError      bool
	IllegalFieldErrors []string
	MissingFieldErrors []string
	InvalidFieldErrors []string
	UniqueErrors       map[string]bool
}

func NewErrorResponse(message string) *ErrorResponse {
	return &ErrorResponse{
		message:       message,
		ContainsError: true,
		UniqueErrors:  make(map[string]bool),
	}
}

func (e *ErrorResponse) SetMessage(m string) {
	e.message = m
	e.ContainsError = true
}

func (e *ErrorResponse) GetAllErrors() []string {
	errorMessages := append(e.IllegalFieldErrors, e.InvalidFieldErrors...)
	return append(errorMessages, e.MissingFieldErrors...)
}

func (e *ErrorResponse) PrintAllErrors() {
	allErrors := e.GetAllErrors()

	if e.ContainsError {
		fmt.Println(e.message)
	}

	for i, e := range allErrors {
		fmt.Println(e)
		if len(allErrors)-1 > i {
			fmt.Println()
		}
	}
}

func (e *ErrorResponse) Contains(key string) bool {
	return e.UniqueErrors[key]
}

func (e *ErrorResponse) FormatValidationError(res *responseErrorItem) {
	// TODO: Structs ? Better usage for edit?
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
			message.Field.Source,
			message.Field.Path,
			message.Field.Value,
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
					message.Field.Source,
					message.Field.Path,
					message.Field.Value,
					message.Message,
				)
				e.IllegalFieldErrors = append(e.IllegalFieldErrors, illegal)
			}

		case "INVALID":
			{
				invalid := fmt.Sprintf(invalidFieldFormat,
					message.Field.Source,
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
		}
	}
}
