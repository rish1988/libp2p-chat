package config

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/rish1988/libp2p-chat/log"
	"os"
	"os/user"
	"path"
	"strings"
)

type Config struct {
	*PrivateKey  `json:"privateKey"`
	*AppConfig   `json:"appConfig"`
	*RemotePeers `json:"remotePeers"`
	ListenAddrs  []string `json:"listenAddrs"`
	Username     string   `json:"name"`
}

type RemotePeers []*RemotePeer

type RemotePeer struct {
	DialAddrs []string `json:"dialAddrs"`
	UserName  string   `json:"name"`
}

func (rp *RemotePeer) Addresses() []peer.AddrInfo {
	var (
		multiAddrs []multiaddr.Multiaddr
	)

	for _, da := range rp.DialAddrs {
		if addr, err := multiaddr.NewMultiaddr(da); err != nil {
			log.Errorln(err)
		} else {
			multiAddrs = append(multiAddrs, addr)
		}
	}

	if multiAddrs != nil {
		if addrInfo, err := peer.AddrInfosFromP2pAddrs(multiAddrs...); err != nil {
			log.Errorln(err)
		} else {
			return addrInfo
		}
	}
	return nil
}

func (rp *RemotePeer) PeerId() *peer.ID {
	addrs := rp.Addresses()
	if addrs != nil {
		return &addrs[0].ID
	}
	return nil
}

type AppConfig struct {
	AppName    string `json:"appName"`
	AppVersion string `json:"appVersion"`
	ProtocolID string `json:"protocolID"`
}

func (a *AppConfig) Protocol() protocol.ID {
	return protocol.ID("/" + a.ProtocolID + "/" + a.AppVersion)
}

type PrivateKey struct {
	Dir      string `json:"dir"`
	FileName string `json:"fileName"`
}

func (p *PrivateKey) IsPresent() bool {
	var (
		privKeyPath = p.privateKeyAbsPath()
		privKeyDir  = path.Dir(*privKeyPath)
		hasPrivKey  bool
	)

	hasPrivKey = len(p.FileName) != 0 && len(p.Dir) != 0

	if !hasPrivKey {
		return false
	}

	if _, err := os.Stat(privKeyDir); os.IsNotExist(err) {
		return false
	} else if _, err := os.Stat(*privKeyPath); os.IsNotExist(err) {
		return false
	}

	return true
}

func (p *PrivateKey) Read() (crypto.PrivKey, error) {
	var (
		privKeyPath = p.privateKeyAbsPath()
		err         error
	)

	if _, err = os.Stat(*privKeyPath); err != nil {
		return nil, err
	} else {
		if contents, err := os.ReadFile(*privKeyPath); err != nil {
			return nil, err
		} else {
			var privKeyDecoded = make([]byte, len(contents)/2)
			if _, err = hex.Decode(privKeyDecoded, contents); err != nil {
				return nil, err
			}
			privKey, _ := btcec.PrivKeyFromBytes(privKeyDecoded)
			return (*crypto.Secp256k1PrivateKey)(privKey), nil
		}
	}
}

func (p *PrivateKey) absPath(rpath string) (*string, error) {
	var absPath string

	if strings.HasPrefix(rpath, "~") {
		if u, err := user.Current(); err != nil {
			return nil, err
		} else {
			absPath = path.Join(u.HomeDir, rpath[1:])
		}
	}
	return &absPath, nil
}

func (p *PrivateKey) privateKeyAbsPath() *string {
	if privKeyPath, err := p.absPath(path.Join(p.Dir, p.FileName)); err != nil {
		return nil
	} else {
		return privKeyPath
	}
}

func (p *PrivateKey) Write(privKey *[]byte) error {
	var (
		privKeyPath = p.privateKeyAbsPath()
		privKeyDir  = path.Dir(*privKeyPath)
	)

	prv, _ := btcec.PrivKeyFromBytes(*privKey)

	prvAsHex := hex.EncodeToString(prv.Serialize())

	if _, err := os.Stat(privKeyDir); os.IsNotExist(err) {
		log.Infof("Creating directory %s", privKeyDir)
		if err := os.MkdirAll(privKeyDir, 0700); err != nil {
			return err
		} else {
			return os.WriteFile(*p.privateKeyAbsPath(), []byte(prvAsHex), 0600)
		}
	}

	return os.WriteFile(*p.privateKeyAbsPath(), []byte(prvAsHex), 0600)
}

func (p *PrivateKey) GenerateSecp256k1KeyPair() (crypto.PrivKey, crypto.PubKey) {
	priv, pub, err := crypto.GenerateKeyPair(crypto.Secp256k1, 256)

	if err != nil {
		log.Errorf("Could not generate keys %v\n", err)
	}
	return priv, pub
}

var config = &Config{}

func (c *Config) ReadConfig() (*Config, error) {
	configFileName := flag.String("configFile", "~/.libp2p-tutorial/config.json", "Application configuration file")
	flag.Parse()
	return parseConfig(configFileName)
}

func parseConfig(configFileName *string) (*Config, error) {
	if _, err := os.Stat(*configFileName); err == nil {
		log.Infof("Loading config: %v", *configFileName)

		cfgFile, err := os.Open(*configFileName)
		if err != nil {
			log.Errorf("File error: %v", err.Error())
			return nil, err
		}
		defer func() {
			err = cfgFile.Close()

			if err != nil {
				log.Errorf("Could not close config file: %v  error: %v", cfgFile.Name(), err)
			}
		}()
		jsonParser := json.NewDecoder(cfgFile)
		if err = jsonParser.Decode(config); err != nil {
			log.Errorf("ReadConfig could not parse json error %v  \n file: %v", err, cfgFile.Name())
			return nil, err
		}
	} else if os.IsNotExist(err) {
		printUsage()
		return nil, err
	}

	return config, nil
}

func printUsage() {
	fmt.Printf(`Libp2p Application %s
Usage:  libp2p-tutorial -configFile=<config-filepath>
		
-config-filepath
	(Required) Specifies where to find the configuration file - config.json. The configuration file must follow JSON specification.

Examples:

$ libp2p-tutorial -configFile=/Users/rishab/Projects/development/go/src/github.com/rish1988/libp2p-tutorial/config/config.json
`, "v0.1-alpha")
}
