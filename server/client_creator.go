package server

import (
	"sync"

	abcicli "github.com/hdac-io/tendermint/abci/client"
	"github.com/hdac-io/tendermint/abci/types"
	"github.com/hdac-io/tendermint/proxy"
)

//----------------------------------------------------
// local proxy uses a mutex on an in-proc app

type fridayLocalClientCreator struct {
	mtx *sync.Mutex
	app types.Application
}

func NewFridayLocalClientCreator(app types.Application) proxy.ClientCreator {
	return &fridayLocalClientCreator{
		mtx: new(sync.Mutex),
		app: app,
	}
}

func (l *fridayLocalClientCreator) NewABCIClient() (abcicli.Client, error) {
	return abcicli.NewFridayLocalClient(l.mtx, l.app), nil
}
