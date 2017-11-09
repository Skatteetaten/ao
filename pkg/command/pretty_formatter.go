package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

type PrettyFormatter struct{}

func (df *PrettyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	level := strings.ToUpper(entry.Level.String())
	time := entry.Time.Format("02-01-2006 15:04:05 CET")
	b.WriteString(fmt.Sprintf("%s %s ", level, time))

	jsonStart := strings.Index(entry.Message, "{")

	if jsonStart >= 0 {
		var pretty bytes.Buffer
		err := json.Indent(&pretty, []byte(entry.Message[jsonStart:]), "", "  ")
		if err == nil {
			b.WriteString(entry.Message[:jsonStart])
			b.WriteString("\n")
			b.Write(pretty.Bytes())
		} else {
			b.WriteString(entry.Message)
		}
	} else {
		b.WriteString(entry.Message)
	}
	if len(entry.Data) == 0 {
		return append(b.Bytes(), '\n'), nil
	}
	serialized, err := json.MarshalIndent(entry.Data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	b.WriteString("\n")
	b.Write(serialized)
	return append(b.Bytes(), '\n'), nil
}
