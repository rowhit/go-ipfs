package iface

import (
	"context"

	"github.com/ipfs/go-ipfs/core/coreapi/interface/options"

	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
)

// DhtAPI specifies the interface to the DHT
type DhtAPI interface {
	// FindPeer queries the DHT for all of the multiaddresses associated with a
	// Peer ID
	FindPeer(context.Context, peer.ID) (pstore.PeerInfo, error)

	// FindProviders finds peers in the DHT who can provide a specific value
	// given a key.
	FindProviders(context.Context, Path, ...options.DhtFindProvidersOption) (<-chan pstore.PeerInfo, error)

	// Provide announces to the network that you are providing given values
	Provide(context.Context, Path, ...options.DhtProvideOption) error
}
