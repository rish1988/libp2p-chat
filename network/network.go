package network

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/routing"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/mr-tron/base58"
	"github.com/rish1988/libp2p-chat/chat"
	"github.com/rish1988/libp2p-chat/config"
	"github.com/rish1988/libp2p-chat/log"
	"golang.org/x/net/context"
	"strings"
)

func BootStrapApp(privKey *crypto.PrivKey, cfg *config.Config) *host.Host {
	ctx := context.Background()

	err, protobufEncodedPubKey := printPubKeyInfo(privKey)

	var idht *dht.IpfsDHT

	//_, err = tor.NewBuilder(
	//	tcfg.AllowTcpDial,                 // Some Configurator are already ready to use.
	//	tcfg.SetSetupTimeout(time.Minute), // Some require a parameter, in this case it's a function that will return a Configurator.
	//	tcfg.SetBinaryPath("/opt/homebrew/bin/tor"),
	//)

	if err != nil {
		log.Error(err)
		return nil
	}

	//key, _ := (*privKey).Raw()
	//var dest []byte
	//base32.HexEncoding.Encode(dest, key)
	//base32Addr := fmt.Sprintf("/onion3/%s:5005", key)
	//cfg.ListenAddrs = append(cfg.ListenAddrs, base32Addr)

	node, err := libp2p.New(
		libp2p.Identity(*privKey),
		libp2p.ListenAddrStrings(cfg.ListenAddrs...),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		//libp2p.DefaultTransports,
		libp2p.NATPortMap(),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			log.Infoln("Creating Kademlia Hash table (KHT)")
			idht, err = dht.New(ctx, h)
			return idht, err
		}),
		libp2p.EnableRelay(),
		libp2p.EnableRelayService(),
		libp2p.EnableHolePunching(),
		//libp2p.EnableAutoRelay(),
		// If you want to help other peers to figure out if they are behind
		// NATs, you can launch the server-side of AutoNAT too (AutoRelay
		// already runs the client)
		libp2p.EnableNATService(),
		//libp2p.Transport(tpt),
	)

	if err != nil {
		log.Errorln(err)
	}

	peerInfo := peer.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}

	printAddrInfo(err, peerInfo)

	printPeerIdInfo(protobufEncodedPubKey, node)

	if cfg.RemotePeers == nil || len(*cfg.RemotePeers) == 0 {
		responder(cfg, &node)
	} else {
		initiator(cfg, &node, ctx)
	}

	return &node
}

func initiator(cfg *config.Config, node *host.Host, ctx context.Context) {
	for _, rp := range *cfg.RemotePeers {
		for _, addr := range rp.Addresses() {
			(*node).Peerstore().AddAddrs(*rp.PeerId(), addr.Addrs, peerstore.PermanentAddrTTL)
		}
		chat.NewChat(ctx, node, rp, cfg)
	}
}

func responder(cfg *config.Config, node *host.Host) {
	log.Infof("Bootstrapping stream handler for %v protocol", cfg.Protocol())
	ch := &chat.Chat{Username: &cfg.Username, Host: node}
	ch.WelcomeMessage(&cfg.Username)
	(*node).SetStreamHandler(cfg.Protocol(), ch.StreamHandler)
}

func printPubKeyInfo(privKey *crypto.PrivKey) (error, []byte) {
	privKeyBytes, err := (*privKey).Raw()

	if err != nil {
		log.Errorln(err)
	}

	_, pub := btcec.PrivKeyFromBytes(privKeyBytes)

	log.Infoln(strings.Repeat("=", 80))
	log.Infoln("Pubkey Info")
	log.Infoln(strings.Repeat("=", 80))

	log.Infof("Using public key (x,y): (%v,%v)", hex.EncodeToString(pub.X().Bytes()), hex.EncodeToString(pub.Y().Bytes()))
	log.Infof("Using public key (Uncompressed): %v", hex.EncodeToString(pub.SerializeUncompressed()))
	log.Infof("Using public key (Compressed): %v", hex.EncodeToString(pub.SerializeCompressed()))
	protobufEncodedPubKey, _ := (*privKey).GetPublic().Raw()
	log.Infof("Pubkey protobuf: <KeyType=0x02 (Secp256k1)><Pubkey data=0x%v>", hex.EncodeToString(pub.SerializeCompressed()))
	log.Infoln("Protobuf Encoding: 0x08 (WireType=VarInt (0), Field number = 1) <KeyType> 0x12 (WireType=Bytes(2), Field number = 2)<Bytes Length=0x21><Pubkey data>")
	log.Infof("Using public key (Protobuf Encoded): %v", hex.EncodeToString(protobufEncodedPubKey))

	log.Infoln(strings.Repeat("=", 80))
	return err, protobufEncodedPubKey
}

func printAddrInfo(err error, peerInfo peer.AddrInfo) {
	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)

	log.Infoln(strings.Repeat("=", 80))
	log.Infoln("Listening on addresses")
	log.Infoln(strings.Repeat("=", 80))

	for _, a := range addrs {
		log.Infoln(a.String())
	}

	log.Infoln(strings.Repeat("=", 80))
}

func connectToBootstrapPeers(h *host.Host, peerAddresses []*config.RemotePeer) error {
	for _, peerAddress := range peerAddresses {
		peerAddrs := peerAddress.Addresses()
		for _, addr := range peerAddrs {
			if err := (*h).Connect(context.Background(), addr); err != nil {
				log.Errorln("Connection failed:", err)
			} else {
				log.Infoln("Connected to:", addr.ID)
			}
		}
	}
	return nil
}

func printPeerIdInfo(protobufEncodedPubKey []byte, node host.Host) {
	ps := node.Peerstore()
	addrs := ps.PeersWithAddrs()
	for _, addr := range addrs {
		log.Infoln(addr)
	}
	log.Infoln("Node PeerID: <Identity=0x00><Length-Pubkey-Protobuf=0x25><Pubkey-Protobuf-Encoded>")
	log.Infof("Node PeerID: 0025%v", hex.EncodeToString(protobufEncodedPubKey))
	log.Infof("Node PeerID (base58btc): %v", base58.Encode(append([]byte{0x00, 0x25}, protobufEncodedPubKey...)))
	log.Infof("Node PeerID (libp2p-base58btc): %v", node.ID())
}
