package main

import (
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	_ "github.com/libp2p/go-libp2p/p2p/protocol/identify"
	"github.com/rish1988/libp2p-chat/config"
	"github.com/rish1988/libp2p-chat/log"
	"github.com/rish1988/libp2p-chat/network"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := (&config.Config{}).ReadConfig()

	if err != nil {
		os.Exit(1)
	}

	var (
		node    *host.Host
		privKey crypto.PrivKey
	)

	if cfg.PrivateKey == nil || !cfg.IsPresent() {
		log.Infoln("Generating Private key")
		privKey, _ = cfg.GenerateSecp256k1KeyPair()
		if privKeyBytes, err := privKey.Raw(); err != nil {
			log.Errorf("Could not export private key %v\n", err)
		} else if err := cfg.Write(&privKeyBytes); err != nil {
			log.Errorf("Could not write private key %v\n", err)
		}
	} else if privKey, err = cfg.Read(); err != nil {
		log.Errorf("Could not read configured private key %v", err)
		os.Exit(1)
	}

	node = network.BootStrapApp(&privKey, cfg)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Infof("\nReceived signal, shutting down...")

	// shut the node down
	if err := (*node).Close(); err != nil {
		log.Errorln(err)
	}
}
