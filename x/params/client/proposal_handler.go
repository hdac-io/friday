package client

import (
	govclient "github.com/hdac-io/friday/x/gov/client"
	"github.com/hdac-io/friday/x/params/client/cli"
	"github.com/hdac-io/friday/x/params/client/rest"
)

// param change proposal handler
var ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
