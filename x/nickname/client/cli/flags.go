package cli

import (
	"os"
)

const (
	FlagAddress = "address"
	FlagWallet  = "wallet"
)

var (
	DefaultClientHome = os.ExpandEnv("$HOME/.clif")
)