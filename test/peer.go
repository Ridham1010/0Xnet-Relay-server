package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
	"github.com/multiformats/go-multiaddr"
)

func main() {
	relayAddr := flag.String("relay", "", "Relay server multiaddr")
	name := flag.String("name", "test-peer", "Peer name")
	flag.Parse()

	if *relayAddr == "" {
		fmt.Println("Usage: go run peer.go -relay <relay-multiaddr>")
		os.Exit(1)
	}

	fmt.Printf("🚀 Starting peer: %s\n", *name)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	privKey, _, _ := crypto.GenerateEd25519Key(rand.Reader)
	relayMA, _ := multiaddr.NewMultiaddr(*relayAddr)
	relayInfo, _ := peer.AddrInfoFromP2pAddr(relayMA)

	h, err := libp2p.New(
		libp2p.ChainOptions(
			libp2p.Transport(tcp.NewTCPTransport),
			libp2p.Transport(websocket.New),
		),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.Identity(privKey),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(noise.ID, noise.New),
		libp2p.EnableRelay(),
	)
	if err != nil { panic(err) }

	fmt.Printf("✅ Peer ID: %s\n", h.ID())

	if err := h.Connect(ctx, *relayInfo); err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		os.Exit(1)
	}

	_, err = client.Reserve(ctx, h, *relayInfo)
	if err != nil {
		fmt.Printf("❌ Reservation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Connected! Relayed Address:\n%s/p2p-circuit/p2p/%s\n", *relayAddr, h.ID())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}