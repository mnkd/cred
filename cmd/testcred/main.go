// testcred prints the values of specified environment variables to stdout.
// Used as a target command for cred run integration tests.
//
// Usage: testcred KEY1 KEY2 ...
// Output: KEY1=value1\nKEY2=value2\n
package main

import (
	"fmt"
	"os"
)

func main() {
	for _, key := range os.Args[1:] {
		fmt.Printf("%s=%s\n", key, os.Getenv(key))
	}
}
