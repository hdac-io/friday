package cli

import (
	"os"
)

var (
	DefaultClientHome = os.ExpandEnv("$HOME/.clif")
)
