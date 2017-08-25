package printutil

import (
	"fmt"
	"strconv"
)

func PrintTable(headers []string, columnValues ...[]string) (output string) {
	columns := len(columnValues)
	fmt.Println("DEBUG: columns: " + strconv.Itoa(columns))
	for _, v := range columnValues {
		fmt.Print(v[0] + " - ")
	}

	// Find max rownum in table

	/*
		// var outputWriter io.Writer
		outputBuffer := new(bytes.Buffer)

		w := tabwriter.NewWriter(outputBuffer, 0, 5, 5, ' ', 0)
		fmt.Fprintln(w, "ENVIRONEMENT\tAPPLICATION")
		// Loop and print
		var lineIndex int = 0

		rows := len(fuzzyArgs.appList)
		if len(fuzzyArgs.envList) > rows {
			rows = len(fuzzyArgs.envList)
		}

		for lineIndex < rows {

			if lineIndex < len(fuzzyArgs.envList) {
				fmt.Fprint(w, fuzzyArgs.envList[lineIndex])
			}
			fmt.Fprint(w, "\t")
			if lineIndex < len(fuzzyArgs.appList) {
				fmt.Fprint(w, fuzzyArgs.appList[lineIndex])
			}
			fmt.Fprintln(w, "")
			lineIndex++
		}
		w.Flush()

		output += "\n" + outputBuffer.String()
	*/
	return
}
