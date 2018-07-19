package coreapi

import (
	"context"
	"errors"
	"fmt"

	coreiface "github.com/ipfs/go-ipfs/core/coreapi/interface"
	caopts "github.com/ipfs/go-ipfs/core/coreapi/interface/options"
	dag "github.com/ipfs/go-ipfs/merkledag"
	"github.com/ipfs/go-ipfs/blockservice"

	"gx/ipfs/QmZ383TySJVeZWzGnWui6pRcKyYZk9VkKTuW7tmKRWk5au/go-libp2p-routing"
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	"gx/ipfs/QmRatnbGjPcoyzVjfixMZnuT1xQbjM7FgnL6FX4CKJeDE2/go-ipfs-blockstore"
	"gx/ipfs/QmShbyKV9P7QuFecDHXsgrQ4rxxm71MUkGVpwedT4VQ8Bf/go-ipfs-exchange-offline"
)

type DhtAPI struct {
	*CoreAPI
	*caopts.DhtOptions
}

func (api *DhtAPI) FindPeer(ctx context.Context, p peer.ID) (pstore.PeerInfo, error) {
	pi, err := api.node.Routing.FindPeer(ctx, peer.ID(p))
	if err != nil {
		return pstore.PeerInfo{}, err
	}

	return pi, nil
}

func (api *DhtAPI) FindProviders(ctx context.Context, p coreiface.Path, opts ...caopts.DhtFindProvidersOption) (<-chan pstore.PeerInfo, error) {
	settings, err := caopts.DhtFindProvidersOptions(opts...)
	if err != nil {
		return nil, err
	}

	rp, err := api.ResolvePath(ctx, p)
	if err != nil {
		return nil, err
	}

	numProviders := settings.NumProviders
	if numProviders < 1 {
		return nil, fmt.Errorf("number of providers must be greater than 0")
	}

	pchan := api.node.Routing.FindProvidersAsync(ctx, rp.Cid(), numProviders)
	return pchan, nil
}

func (api *DhtAPI) Provide(ctx context.Context, path coreiface.Path, opts ...caopts.DhtProvideOption) error {
	settings, err := caopts.DhtProvideOptions(opts...)
	if err != nil {
		return err
	}

	if api.node.Routing == nil {
		return errors.New("cannot provide in offline mode")
	}

	rp, err := api.ResolvePath(ctx, path)
	if err != nil {
		return err
	}

	c := rp.Cid()

	has, err := api.node.Blockstore.Has(c)
	if err != nil {
		return err
	}

	if !has {
		return fmt.Errorf("block %s not found locally, cannot provide", c)
	}

	if settings.Recursive {
		err = provideKeysRec(ctx, api.node.Routing, api.node.Blockstore, []*cid.Cid{c})
	} else {
		err = provideKeys(ctx, api.node.Routing, []*cid.Cid{c})
	}
	if err != nil {
		return err
	}

	return nil
}

func provideKeys(ctx context.Context, r routing.IpfsRouting, cids []*cid.Cid) error {
	for _, c := range cids {
		err := r.Provide(ctx, c, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func provideKeysRec(ctx context.Context, r routing.IpfsRouting, bs blockstore.Blockstore, cids []*cid.Cid) error {
	provided := cid.NewSet()
	for _, c := range cids {
		kset := cid.NewSet()

		dserv := dag.NewDAGService(blockservice.New(bs, offline.Exchange(bs)))

		err := dag.EnumerateChildrenAsync(ctx, dag.GetLinksDirect(dserv), c, kset.Visit)
		if err != nil {
			return err
		}

		for _, k := range kset.Keys() {
			if provided.Has(k) {
				continue
			}

			err = r.Provide(ctx, k, true)
			if err != nil {
				return err
			}
			provided.Add(k)
		}
	}

	return nil
}
