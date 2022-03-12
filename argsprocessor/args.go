package argsprocessor

import (
	"io"
	"os"
)

// ProcessArgs reads either stdin or a series of files whose names are in the argument list specified by args
func ProcessArgs(args []string, stdin *os.File, stderr io.Writer, eachFile func(filename string, file *os.File)) {
	switch len(args) {
	case 0:
		// Use stdin
		eachFile("", stdin)
	default:
		for _, filename := range args {
			f, err := os.Open(filename)
			if err != nil {
				stderr.Write([]byte(err.Error()))
				stderr.Write([]byte("\n"))
			} else {
				defer f.Close()
				eachFile(filename, f)
			}
		}
	}
}
