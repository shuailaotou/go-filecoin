package impl

import (
	"context"

	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"

	api "github.com/filecoin-project/go-filecoin/api"
)

type nodeStats struct {
	api *nodeAPI
}

func newNodeStats(api *nodeAPI) *nodeStats {
	return &nodeStats{api: api}
}

// Get returns the associated DAG node for the passed in CID.
func (api *nodeStats) Bandwidth(ctx context.Context, proto string) api.BWStats {
	if proto != "" {
		return api.api.node.BandwidthTracker.GetBandwidthForProtocol(protocol.ID(proto))
	}

	return api.api.node.BandwidthTracker.GetBandwidthTotals()
}
