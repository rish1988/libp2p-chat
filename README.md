# Libp2p-Chat

Peer to peer chatting application build on top of Libp2p. The chatting app comes
with two default configs - `config.initiator.json` and `config.responder.json` which are
used by libp2p to bootstrap initiator and responder nodes.

This chatting application works as long as the initiator can dial the `dialAddrs` in its relevant config.

What the initiator and responder nodes first run, the relevant private keys are generated and stored in `~/.libp2p-chat/initiator/libp2p.prv` and
`~/.libp2p-chat/responder/libp2p.prv`. On subsequent runs, these keys are reused for bootstrapping the node.
The directory and private key name for storing the generated private keys can be configured via
`privateKey.dir` and `privateKey.fileName` in relevant config files.

The current state of the application doesn't support node discovery. Feel free to add peer discovery to the application.

# Prerequisites

* Go (version 1.22 or higher)
* Go mod feature turned on

# Building

To build the project simply execute

```
# Builds the Go module
$ go build

# Adds the chatting app to Path
$ mv libp2p-chat /usr/local/bin
```

# Running the Chat Application

## Start Responder (Bob) Node

* Start responder (Bob's node) by executing the below command

```bash
# The configFile switch configures the path to your responder config file 
$ libp2p-chat -configFile=<path-to-config-responder-json>
```

After the Bob's node is bootstrapped, you should see something like this

```
=================================
Hello Bob!
Welcome to Chatting Application
Instructions: Type a message and press enter to send it
=================================
Bob>
```

## Start Initiator (Alice) Node

Start initiator (Alice's node) by executing the below command

```bash
# The configFile switch configures the path to your initiator config file 
$ libp2p-chat -configFile=<path-to-config-initiator-json>
```

After the Alice's node is bootstrapped, you should see something like this


```
=================================
Hello Alice!
Welcome to Chatting Application
Instructions: Type a message and press enter to send it
=================================
Alice>
```

Make sure to update the `peerId` of remote peer's multiaddress configured in `dialAddrs`
after initiator node spits out its `PeerId`

## Package Info

The application consists of the following packages

* `chat` - Implements the chat application protocol
* `log` - Enables logs with colors
* `config` - Reads configuration related to the Node and provides two default configs `config.initiator.json` and `config.responder.json`
* `network` - Bootstraps the responder or initiator node with config read by `config` package and wires
             the chat application protocol handlers to the network stream after initial protocol negotiation
             done by libp2p