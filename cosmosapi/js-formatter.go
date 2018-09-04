package cosmosapi

import (
	"regexp"
	"strings"
)

func EscapeJavaScript(source []byte) string {
	sourceCode := string(source)

	reReplaceNewLines := regexp.MustCompile(`\r?\n`)
	sourceCode = reReplaceNewLines.ReplaceAllString(sourceCode, "\\n")

	sourceCode = strings.Replace(sourceCode, `"`, `\"`, -1)

	//fmt.Fprintf(os.Stdout, sourceCode)
	return sourceCode
}
