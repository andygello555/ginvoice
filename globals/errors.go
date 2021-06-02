package globals

import (
	"flag"
	"fmt"
	"os"
)

// CliError type for handling errors which occur in the CLI.
type CliError struct {
	// The return code.
	code     int
	// Whether or not the error is internal or down to user input.
	internal bool
	// The message to print along with the err message if given.
	message  string
}

// CliError(s) (positive codes).
var (
	// ParseErrUser occurs when a parse error is down to malformed user input.
	ParseErrUser         = CliError{1, false, "The following value cannot be parsed"}
	RequiredFlag         = CliError{2, false, "The following flags are required"}
	FileErrUser          = CliError{3, false, "A user file error occurred"}
	FileErr              = CliError{4, true, "A file error occurred"}
	InvoiceGenerationErr = CliError{5, true, "Error when generating invoice"}
)

// Handle the print of the error details as well as exiting with the defined exit code.
func (e *CliError) Handle(err error) {
	if err != nil {
		fmt.Println(e.message + ":", err)
	} else {
		fmt.Println(e.message)
	}

	// PrintDefaults if not an internal error
	if !e.internal {
		flag.PrintDefaults()
	}

	// Finally exit, returning the exit code to the shell
	os.Exit(e.code)
}
