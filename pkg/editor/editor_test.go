package editor

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHasContentChanged(t *testing.T) {

	testCases := []struct {
		Original string
		Edited   string
		Expected bool
	}{
		{
			Original: `{ "type": "development" }`,
			// Whitespace
			Edited:   `{     "type": "development"          }`,
			Expected: false,
		},
		{
			Original: `{ "type": "development" }`,
			// Newline
			Edited: `{     "type": "development"
			}`,
			Expected: false,
		},
		{
			Original: `{ "type": "deploy" }`,
			// Changed
			Edited:   `{     "type": "development"          }`,
			Expected: true,
		},
		{
			Original: `{ "type": "deploy" }`,
			// Illegal json
			Edited:   `{type": "deploy"}`,
			Expected: true,
		},
	}

	for _, test := range testCases {
		assert.Equal(t, test.Expected, hasContentChanged(test.Original, test.Edited))
	}

}
