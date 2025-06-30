package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/urfave/cli/v3"
)

func getHostAddress(ha host.Host) string {
	// Build host multiaddress
	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", ha.ID()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addr := ha.Addrs()[0]
	return addr.Encapsulate(hostAddr).String()
}

var app = cli.Command{
	Name: "p2p-tunnel",
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "allowed-peers",
			Usage: "Allowed peers for incoming connections",
		},
		&cli.StringFlag{
			Name:  "identity-path",
			Usage: "Identity path",
			Value: "identity",
		},
		&cli.StringFlag{
			Name:  "listen",
			Usage: "Listen address",
			Value: "/ip4/0.0.0.0/tcp/9999",
		},
		&cli.StringMapFlag{
			Name:  "forward",
			Usage: "Service to forward the data, value is the name of the P2P service endpoint",
		},
		&cli.StringMapFlag{
			Name:  "bind",
			Usage: "Accept local connections and forward them to remote P2P endpoints",
		},
	},

	Action: func(ctx context.Context, c *cli.Command) (err error) {
		privKey, err := LoadIdentity(c.String("identity-path"))

		host, err := libp2p.New(
			libp2p.ListenAddrStrings(c.String("listen")),
			libp2p.Identity(privKey),
		)
		if err != nil {
			return fmt.Errorf("failed to create host: %w", err)
		}
		defer host.Close()

		log.Println("[+] Listening")
		log.Println("\tId:", host.ID())
		forward := c.StringMap("forward")
		bind := c.StringMap("bind")
		maddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", host.ID()))

		log.Println("\t[+] Forward")
		for target, service := range forward {
			for _, addr := range host.Addrs() {
				log.Println("\t", addr.Encapsulate(maddr).String()+"/"+service, "->", target)
			}
		}

		log.Println("\t[+] Bind")
		for target, service := range bind {
			for _, addr := range host.Addrs() {
				log.Println("\t", addr.Encapsulate(maddr).String()+"/"+service, "->", target)
			}
		}

		log.Println("[*] Preparing forward listeners")
		allowedPeers := map[string]struct{}{}
		for _, allowedPeer := range c.StringSlice("allowed-peers") {
			log.Println("[*] Permitting peer:", allowedPeer)
			allowedPeers[allowedPeer] = struct{}{}
		}

		for target, service := range forward {
			log.Println("\t[*] Preparing:", target, "over protocol:", service)
			host.SetStreamHandler(protocol.ID(service), func(s network.Stream) {
				remotePeer := s.Conn().RemotePeer()
				log.Println("[*] Received connection from peer:", remotePeer)
				defer s.Close()

				if len(allowedPeers) > 0 {
					_, found := allowedPeers[remotePeer.String()]
					if !found {
						log.Println("[!] Blocked:", remotePeer)
						return
					}
				}

				log.Println("[*] Connecting to target:", target)
				conn, err := net.Dial("tcp", target)
				if err != nil {
					log.Println("[!] Failed to dial to:", target)
					return
				}

				log.Println("[*] Forwarding data:", target)
				go io.Copy(conn, s)
				io.Copy(s, conn)
			})
		}

		log.Println("[*] Preparing bind listeners")
		var listeners []net.Listener
		defer func() {
			for _, l := range listeners {
				l.Close()
			}
		}()
		for src, target := range bind {
			split := strings.Split(target, ":")
			if len(split) != 2 {
				return fmt.Errorf("expecting ':' pointing to the service name for: %s", target)
			}

			rawPeerAddr := split[0]
			service := split[1]
			log.Println("\t[*] Preparing:", src, "over peer:", rawPeerAddr)
			var peerAddr multiaddr.Multiaddr
			peerAddr, err = multiaddr.NewMultiaddr(rawPeerAddr)
			if err != nil {
				return fmt.Errorf("failed to parse peer address: %w", err)
			}

			var info *peer.AddrInfo
			info, err = peer.AddrInfoFromP2pAddr(peerAddr)
			if err != nil {
				return fmt.Errorf("failed to get peer info: %w", err)
			}

			host.
				Peerstore().
				AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

			var l net.Listener
			l, err = net.Listen("tcp", src)
			if err != nil {
				return fmt.Errorf("failed to listen at target: %w", err)
			}
			listeners = append(listeners, l)

			go func() {
				for {
					conn, err := l.Accept()
					if err != nil {
						log.Println(err)
						return
					}

					go func() {
						defer conn.Close()
						log.Println("[*] Received connection:", conn.RemoteAddr(), "for service:", target)

						log.Println("[*] Connecting to peer:", rawPeerAddr, "for service:", target)
						s, err := host.NewStream(context.Background(), info.ID, protocol.ID(service))
						if err != nil {
							log.Println(err)
							return
						}
						defer s.Close()

						defer log.Println("[*] Served request:", rawPeerAddr, "for service:", target)
						log.Println("[*] Serving request:", rawPeerAddr, "for service:", target)
						go io.Copy(s, conn)
						io.Copy(conn, s)
					}()
				}
			}()
		}

		<-ctx.Done()
		return ctx.Err()
	},
}

func main() {
	ctx := context.TODO()
	err := app.Run(ctx, os.Args)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[+] Completed successfully")
}
