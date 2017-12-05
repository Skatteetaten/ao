package log

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPrettyFormatter_Format(t *testing.T) {

	cases := []struct {
		Message string
		Fields  logrus.Fields
		// Exclude log level and time stamp
		Expected string
	}{
		{
			"Hello",
			logrus.Fields{"status": "ok"},
			"Hello\n{\n  \"status\": \"ok\"\n}\n",
		},
		{
			"Hello{{",
			logrus.Fields{},
			"Hello{{\n",
		},
		{
			`Response {"name":"foo"}`,
			logrus.Fields{},
			"Response \n{\n  \"name\": \"foo\"\n}\n",
		},
		{
			`Response {"name":"foo"}`,
			logrus.Fields{"status": "fail"},
			"Response \n{\n  \"name\": \"foo\"\n}\n{\n  \"status\": \"fail\"\n}\n",
		},
	}

	now := time.Date(2017, 11, 22, 15, 0, 0, 0, time.Local)

	for _, tc := range cases {
		entry := &logrus.Entry{
			Message: tc.Message,
			Data:    tc.Fields,
			Level:   logrus.InfoLevel,
			Time:    now,
		}

		pretty := PrettyFormatter{}
		result, err := pretty.Format(entry)
		if err != nil {
			t.Fatal(err)
		}

		expected := "INFO Wed, 22 Nov 2017 15:00:00 CET " + tc.Expected
		assert.Equal(t, expected, string(result))
	}
}
