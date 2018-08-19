package main

import (
	"fmt"
	"github.com/vippsas/go-cosmosdb"
	"io/ioutil"
	"os"
)

// Format a JavaScript-file for inline use in a JSON file.
// Usage: cat some-script.js | [this cmd] ( | clipboard)
func main() {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	sourceCode := cosmosdb.EscapeJavaScript(bytes)

	fmt.Fprintf(os.Stdout, sourceCode)
}
