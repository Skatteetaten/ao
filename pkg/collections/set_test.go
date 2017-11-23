package collections

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewStringSet(t *testing.T) {

	cases := []struct {
		Items    []string
		Expected []string
	}{
		{[]string{"foo", "foo", "bar", "bar"}, []string{"foo", "bar"}},
	}

	for _, tc := range cases {
		set := NewStringSet()
		for _, item := range tc.Items {
			set.Add(item)
		}
		for _, exp := range tc.Expected {
			assert.Equal(t, true, set.set[exp])
		}

		assert.Equal(t, len(tc.Expected), len(set.All()))
	}
}
