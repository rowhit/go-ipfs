package p2p

import (
	"context"
	"time"

	"gx/ipfs/QmNqRnejxJxjRroz7buhrjfU8i3yNBLa81hFtmf2pXEffN/go-multiaddr-net"
	ma "gx/ipfs/QmUxSEGbv2nmYNnfXi7839wwQqTN3kwQeUxe8dTjZWZs7J/go-multiaddr"
	"gx/ipfs/QmVf8hTAsLLFtn4WPCRNdnaF2Eag2qTBS6uR8AiHPZARXy/go-libp2p-peer"
	tec "gx/ipfs/QmWHgLqrghM9zw77nF6gdvT9ExQ2RB9pLxkd8sDHZf1rWb/go-temp-err-catcher"
	"gx/ipfs/QmXdgNhVEgjLxjUoMs5ViQL7pboAt3Y7V7eGHRiE4qrmTE/go-libp2p-net"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
)

// localListener manet streams and proxies them to libp2p services
type localListener struct {
	ctx context.Context

	p2p *P2P
	id  peer.ID

	proto protocol.ID
	laddr ma.Multiaddr
	peer  peer.ID

	listener manet.Listener
}

// ForwardLocal creates new P2P stream to a remote listener
func (p2p *P2P) ForwardLocal(ctx context.Context, peer peer.ID, proto protocol.ID, bindAddr ma.Multiaddr) (Listener, error) {
	listener := &localListener{
		ctx: ctx,

		p2p: p2p,
		id:  p2p.identity,

		proto: proto,
		laddr: bindAddr,
		peer:  peer,
	}

	if err := p2p.Listeners.Register(listener); err != nil {
		return nil, err
	}

	go listener.acceptConns()

	return listener, nil
}

func (l *localListener) dial(ctx context.Context) (net.Stream, error) {
	cctx, cancel := context.WithTimeout(ctx, time.Second*30) //TODO: configurable?
	defer cancel()

	return l.p2p.peerHost.NewStream(cctx, l.peer, l.proto)
}

func (l *localListener) acceptConns() {
	for {
		local, err := l.listener.Accept()
		if err != nil {
			if tec.ErrIsTemporary(err) {
				continue
			}
			return
		}

		go l.setupStream(local)
	}
}

func (l *localListener) setupStream(local manet.Conn) {
	remote, err := l.dial(l.ctx)
	if err != nil {
		local.Close()
		log.Warningf("failed to dial to remote %s/%s", l.peer.Pretty(), l.proto)
		return
	}

	stream := &Stream{
		Protocol: l.proto,

		OriginAddr: local.RemoteMultiaddr(),
		TargetAddr: l.TargetAddress(),

		Local:  local,
		Remote: remote,

		Registry: l.p2p.Streams,
	}

	l.p2p.Streams.Register(stream)
	stream.startStreaming()
}

func (l *localListener) start() error {
	maListener, err := manet.Listen(l.laddr)
	if err != nil {
		return err
	}

	l.listener = maListener
	return nil
}

func (l *localListener) Close() error {
	ok, err := l.p2p.Listeners.Deregister(getListenerKey(l))
	if err != nil {
		return err
	}
	if ok {
		l.listener.Close()
		l.listener = nil
	}
	return nil
}

func (l *localListener) Protocol() protocol.ID {
	return l.proto
}

func (l *localListener) ListenAddress() ma.Multiaddr {
	return l.laddr
}

func (l *localListener) TargetAddress() ma.Multiaddr {
	addr, err := ma.NewMultiaddr(maPrefix + l.peer.Pretty())
	if err != nil {
		panic(err)
	}
	return addr
}
