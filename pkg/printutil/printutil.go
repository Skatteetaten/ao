package printutil

import (
	"bytes"
	"fmt"
	"text/tabwriter"
)

func FormatTable(headers []string, columnValues ...[]string) (output string) {

	var maxRowNum = 0
	for _, v := range columnValues {
		if len(v) > maxRowNum {
			maxRowNum = len(v)
		}
	}

	outputBuffer := new(bytes.Buffer)
	w := tabwriter.NewWriter(outputBuffer, 0, 5, 5, ' ', 0)

	// Output headers
	for hi := range headers {
		if hi > 0 {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprint(w, headers[hi])
	}
	fmt.Fprintln(w, "")

	// Output rows
	var lineIndex int = 0
	for lineIndex < maxRowNum {
		for columnIndex := range columnValues {
			if columnIndex > 0 {
				fmt.Fprint(w, "\t")
			}
			if len(columnValues[columnIndex]) > lineIndex {
				fmt.Fprint(w, columnValues[columnIndex][lineIndex])
			}
		}
		lineIndex++
		fmt.Fprintln(w, "")
	}

	w.Flush()
	output = outputBuffer.String()
	//fmt.Println("DEBUG: Len(output): " + strconv.Itoa(len(output)))  // Used to generate test result
	return output
}
