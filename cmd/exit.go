package cmd

import "fmt"

// ExitCodeError reports that cred should exit with Code. It carries the exit
// code of a command started by cred run, so main exits with it without
// printing an error.
type ExitCodeError struct {
	Code int
}

func (e *ExitCodeError) Error() string {
	return fmt.Sprintf("exit status %d", e.Code)
}
