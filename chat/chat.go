package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/rish1988/libp2p-chat/config"
	"github.com/rish1988/libp2p-chat/log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Chat struct {
	Username *string
	Host     *host.Host
	Remote   *config.RemotePeer
}

type chatMessage struct {
	Username string
	Message  string
}

func NewChat(ctx context.Context, h *host.Host, rp *config.RemotePeer, cfg *config.Config) *Chat {
	ch := &Chat{Username: &cfg.Username, Host: h, Remote: rp}
	log.Infof("Remote Peer %s with addresses %v added", *rp.PeerId(), rp.Addresses())
	if s, err := (*h).NewStream(ctx, *rp.PeerId(), cfg.Protocol()); err != nil {
		log.Errorf("Failed to connect to peer")
		log.Errorln(err)
	} else {
		log.Infof("Successfully connected to %s (%s)", rp.UserName, rp.PeerId())
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		go ch.readMessage(rw.Reader, &s)
		go ch.sendMessage(rw.Writer, &s)
	}
	ch.WelcomeMessage(&cfg.Username)
	defer ctx.Done()
	return &Chat{Username: &cfg.Username}
}

func (c *Chat) StreamHandler(s network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	go c.readMessage(rw.Reader, &s)
	go c.sendMessage(rw.Writer, &s)
}

func (c *Chat) readMessage(r *bufio.Reader, s *network.Stream) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(100 * time.Millisecond)
	// If you don't stop it, the ticker will cause memory leaks
	defer ticker.Stop()

	for {
		select {
		case <-ch:
			fmt.Println()
			fmt.Println("\x1b[32mClosing Chat Application\x1b[0m")
			fmt.Printf("\x1b[32mGoodbye %s!\x1b[0m", *c.Username)
			fmt.Println()
			(*s).Close()
			return
		case <-ticker.C:
			var rmessage chatMessage
			if rbytes, err := r.ReadBytes(byte('\n')); err != nil {
				(*s).Close()
				//log.Errorln(err)
			} else if err := json.Unmarshal(rbytes[:len(rbytes)-1], &rmessage); err != nil {
				log.Errorln(err)
			} else {
				if len(rmessage.Message) == 0 {
					return
				} else if rmessage.Message != "\n" {
					fmt.Printf("\r\u001B[32m%s\u001B[0m < %s", rmessage.Username, rmessage.Message)
					fmt.Printf("\x1b[32m%s\x1b[0m > ", *c.Username)
				}
			}
		}
	}
}

func (c *Chat) sendMessage(w *bufio.Writer, s *network.Stream) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		if smessage, err := stdReader.ReadString('\n'); err != nil {
			log.Errorln("Could not read enter message")
			log.Errorln("Please try again")
		} else if smessage, err := json.Marshal(&chatMessage{Username: *c.Username, Message: smessage}); err != nil {
			log.Errorln("Chat application error. Try restarting")
		} else if _, err = w.Write(append(smessage, byte('\n'))); err != nil {
			log.Errorln("Could not send message")
		} else if err = w.Flush(); err != nil {
			log.Errorln("Could not send message")
		}
		fmt.Printf("\x1b[32m%s\x1b[0m > ", *c.Username)
	}
}

func (c *Chat) WelcomeMessage(username *string) {
	log.Infof(strings.Repeat("=", 80))
	log.Infof("Hello %s", *username)
	log.Infof("Welcome to Chatting App")
	log.Infoln("Instructions: Type a message and press enter to send it")
	log.Infof(strings.Repeat("=", 80))
	fmt.Printf("\x1b[32m%s\x1b[0m > ", *c.Username)
}
