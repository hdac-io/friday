package client

import (
	"github.com/gorilla/mux"

	"github.com/hdac-io/friday/client/context"
)

// Register routes
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	RegisterRPCRoutes(cliCtx, r)
}
