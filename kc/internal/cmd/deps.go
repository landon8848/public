package cmd

import (
	"io"

	"github.com/landon8848/public/kc/internal/opclient"
)

// Deps holds the injectable dependencies for kc commands.
type Deps struct {
	RegistryPath string
	OP           *opclient.Client
	In           io.Reader
	Out          io.Writer
	Err          io.Writer
}
