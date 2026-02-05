package main

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
	"github.com/multiformats/go-multiaddr"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h, err := libp2p.New(
		libp2p.ChainOptions(
			libp2p.Transport(tcp.NewTCPTransport),
			libp2p.Transport(websocket.New),
		),
	)
	if err != nil { panic(err) }

	fmt.Println("Client Peer ID:", h.ID())

	relayAddrStr := "/dns4/zeroxnet-relay-server.onrender.com/tcp/443/wss/p2p/12D3KooWRaJtqkLhngjzAfMjiabgKQG1Unu1ouW1ceREs4boytTy"
	relayMA, _ := multiaddr.NewMultiaddr(relayAddrStr)
	relayInfo, _ := peer.AddrInfoFromP2pAddr(relayMA)

	if err := h.Connect(ctx, *relayInfo); err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		return
	}

	_, err = client.Reserve(ctx, h, *relayInfo)
	if err != nil {
		fmt.Printf("❌ Reservation failed: %v\n", err)
		return
	}
	fmt.Println("✅ Client Reserved on Relay!")

	h.SetStreamHandler("/keepalive/1.0.0", func(s network.Stream) {
		fmt.Println("🔗 Stream received!")
		s.Close()
	})

	select {
	case <-ctx.Done():
	case <-time.After(time.Hour):
	}
}