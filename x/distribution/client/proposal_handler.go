package client

import (
	"github.com/hdac-io/friday/x/distribution/client/cli"
	"github.com/hdac-io/friday/x/distribution/client/rest"
	govclient "github.com/hdac-io/friday/x/gov/client"
)

// param change proposal handler
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
)
