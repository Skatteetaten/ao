package client

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

func init() {
	logrus.SetLevel(logrus.FatalLevel)
}

func ReadTestFile(name string) []byte {
	filePath := fmt.Sprintf("./test_files/%s.json", name)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	return data
}
