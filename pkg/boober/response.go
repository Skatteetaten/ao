package boober

import (
	"fmt"
	"strings"
)

type responseErrorItem struct {
	Application string `json:"application"`
	Environment string `json:"environment"`
	Messages []struct {
		Type    string `json:"type"`
		Message string `json:"message"`
		Field struct {
			Path   string `json:"path"`
			Value  string `json:"value"`
			Source string `json:"source"`
		} `json:"field"`
	} `json:"messages"`
}

type responseError struct {
	Response
	Items []responseErrorItem `json:"items"`
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

type Validation struct {
	message            string
	ContainsError bool
	IllegalFieldErrors []string
	MissingFieldErrors []string
	InvalidFieldErrors []string
	UniqueErrors       map[string]bool
}

func NewValidation(message string) *Validation {
	return &Validation{
		message: message,
		ContainsError: true,
		UniqueErrors: make(map[string]bool),
	}
}

func (v *Validation) SetMessage(m string) {
	v.message = m
	v.ContainsError = true
}

func (v *Validation) GetAllErrors() []string {
	errorMessages := append(v.IllegalFieldErrors, v.InvalidFieldErrors...)
	return append(errorMessages, v.MissingFieldErrors...)
}

func (v *Validation) PrintAllErrors() {
	allErrors := v.GetAllErrors()

	if v.ContainsError {
		fmt.Println(v.message)
	}

	for i, e := range allErrors {
		fmt.Println(e)
		if len(allErrors)-1 > i {
			fmt.Println()
		}
	}
}

func (v *Validation) Contains(key string) bool {
	return v.UniqueErrors[key]
}

func (v *Validation) FormatValidationError(res *responseErrorItem) {
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

		if v.Contains(key) {
			continue
		}

		if message.Type != "MISSING" {
			v.UniqueErrors[key] = true
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
				v.IllegalFieldErrors = append(v.IllegalFieldErrors, illegal)
			}

		case "INVALID":
			{
				invalid := fmt.Sprintf(invalidFieldFormat,
					message.Field.Source,
					message.Field.Path,
					message.Message,
				)
				v.InvalidFieldErrors = append(v.InvalidFieldErrors, invalid)
			}

		case "MISSING":
			{
				missing := fmt.Sprintf(missingFieldFormat,
					res.Environment,
					res.Application,
					message.Field.Path,
					message.Message,
				)
				v.MissingFieldErrors = append(v.MissingFieldErrors, missing)
			}
		}
	}
}
