package connectiongating

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"

	"github.com/golang/mock/gomock"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

//go:generate go run github.com/golang/mock/mockgen -package connectiongating -destination mock_connection_gater_test.go github.com/libp2p/go-libp2p/core/connmgr ConnectionGater

// This list should contain (at least) one address for every transport we have.
var addrs = []ma.Multiaddr{
	ma.StringCast("/ip4/127.0.0.1/tcp/0"),
	ma.StringCast("/ip4/127.0.0.1/tcp/0/ws"),
	ma.StringCast("/ip4/127.0.0.1/udp/0/quic"),
	ma.StringCast("/ip4/127.0.0.1/udp/0/quic-v1"),
	ma.StringCast("/ip4/127.0.0.1/udp/0/quic-v1/webtransport"),
}

func TestConnectionGatingInterceptPeerDial(t *testing.T) {
	for _, a := range addrs {
		t.Run(fmt.Sprintf("dialing %s", a), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			connGater := NewMockConnectionGater(ctrl)

			h1, err := libp2p.New(libp2p.ConnectionGater(connGater))
			require.NoError(t, err)
			h2, err := libp2p.New(libp2p.ListenAddrs(a))
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			connGater.EXPECT().InterceptPeerDial(h2.ID())
			require.ErrorIs(t, h1.Connect(ctx, peer.AddrInfo{ID: h2.ID(), Addrs: []ma.Multiaddr{a}}), swarm.ErrGaterDisallowedConnection)
		})
	}
}

func TestConnectionGatingInterceptAddrDial(t *testing.T) {
	for _, a := range addrs {
		t.Run(fmt.Sprintf("dialing %s", a), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			connGater := NewMockConnectionGater(ctrl)

			h1, err := libp2p.New(libp2p.ConnectionGater(connGater))
			require.NoError(t, err)
			h2, err := libp2p.New(libp2p.ListenAddrs(a))
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			gomock.InOrder(
				connGater.EXPECT().InterceptPeerDial(h2.ID()).Return(true),
				connGater.EXPECT().InterceptAddrDial(h2.ID(), a),
			)
			require.ErrorIs(t, h1.Connect(ctx, peer.AddrInfo{ID: h2.ID(), Addrs: []ma.Multiaddr{a}}), swarm.ErrNoGoodAddresses)
		})
	}
}
