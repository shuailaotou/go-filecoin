// Package api holds the interface definitions for the Filecoin api.
package api

import (
	"github.com/filecoin-project/go-filecoin/node/sectorforeman"
)

// API is the user interface to a Filecoin node.
type API interface {
	Actor() Actor
	Address() Address
	Client() Client
	Daemon() Daemon
	Dag() Dag
	ID() ID
	Log() Log
	Miner() Miner
	Mining() Mining
	Paych() Paych
	Ping() Ping
	RetrievalClient() RetrievalClient
	Swarm() Swarm
	SectorForeman() *sectorforeman.SectorForeman
	Version() Version
}
