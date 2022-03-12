package argsprocessor

import (
	"io"
	"os"
)

const STDIN = "-"

// ProcessArgs reads either stdin or a series of files whose names are in the argument list specified by args
func ProcessArgs(args []string, stdin io.Reader, eachFile func(filename string, file io.Reader)) error {
	switch len(args) {
	case 0:
		// Use stdin
		eachFile(STDIN, stdin)
	default:
		for _, filename := range args {
			f, err := os.Open(filename)
			if err != nil {
				return err
			} else {
				defer f.Close()
				eachFile(filename, f)
			}
		}
	}

	return nil
}
