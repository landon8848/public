package cmd

import (
	"fmt"
	"io"
	"os"
)

// readSource returns raw bytes from a file path, or from r when path is "" or "-".
func readSource(path string, r io.Reader) ([]byte, error) {
	if path == "" || path == "-" {
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("no input on stdin")
		}
		return data, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("%s is empty", path)
	}
	return data, nil
}
