package filnet

import (
	"context"

	"gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	pstore "gx/ipfs/QmRhFARzTHcFh8wUxwN5KvyTGq73FLC65EfFAhz8Ng7aGb/go-libp2p-peerstore"
	routing "gx/ipfs/QmWaDSNoSdSXU9b6udyaq9T8y6LkzMwqWxECznFqvtcTsk/go-libp2p-routing"
)

// Router is used to find information about who has what content.
type Router struct {
	routing routing.IpfsRouting
}

// NewRouter builds a new router.
func NewRouter(r routing.IpfsRouting) *Router {
	return &Router{routing: r}
}

// FindProvidersAsync searches for and returns peers who are able to provide a
// given key.
func (r *Router) FindProvidersAsync(ctx context.Context, key cid.Cid, count int) <-chan pstore.PeerInfo {
	return r.routing.FindProvidersAsync(ctx, key, count)
}
