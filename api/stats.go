package api

import (
	"context"

	metrics "gx/ipfs/QmZZseAa9xcK6tT3YpaShNUAEpyRAoWmUL5ojH3uGNepAc/go-libp2p-metrics"
)

// Stats is the interface used to retrieve filecoin node stats
type Stats interface {
	// Bandwidth returns node bandwidth statistics
	Bandwidth(ctx context.Context, proto string) BWStats
}

type BWStats = metrics.Stats
