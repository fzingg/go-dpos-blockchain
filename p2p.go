// This is the p2p network, handler the conn and communicate with nodes each other.


package dpos

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	mrand "math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	host "github.com/libp2p/go-libp2p-host"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/outbrain/golib/log"
	"github.com/urfave/cli"
)

const (

	DefaultVote = 10
	
	FileName = "config.ini"
)

var mutex = &sync.Mutex{}


type Validator struct {
	name string
	vote int
}


var NewNode = cli.Command{
	Name:  "new",
	Usage: "add a new node to p2p network",
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "port",
			Value: 3000,
			Usage: "",
		},
		cli.StringFlag{
			Name:  "target",
			Value: "",
			Usage: "",
		},
		cli.BoolFlag{
			Name:  "secio",
			Usage: "",
		},
		cli.Int64Flag{
			Name:  "seed",
			Value: 0,
			Usage: "",
		},
	},
	Action: func(context *cli.Context) error {
		if err := Run(context); err != nil {
			return err
		}
		return nil
	},
}

func MakeBasicHost(listenPort int, secio bool, randseed int64) (host.Host, error) {
	var r io.Reader

	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}


	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}

	if !secio {
		opts = append(opts, libp2p.NoSecurity)
	}
	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses;
	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)

	log.Infof("I am: %s\n", fullAddr)
	SavePeer(basicHost.ID().Pretty())

	if secio {
		fmt.Printf("'./dpos new --port %d --target %s -secio' \n", listenPort+1, fullAddr)
	} else {
		fmt.Printf("'./dpos new --port %d --target %s' \n", listenPort+1, fullAddr)
	}
	return basicHost, nil
}

// HandleStream  handler stream info
func HandleStream(s network.Stream) {
	log.Infof(" %s", s.Conn().RemotePeer().Pretty())
	
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	go readData(rw)
	go writeData(rw)
}


func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Errorf(err.Error())
		}

		if str == "" {
			return
		}
		if str != "\n" {
			chain := make([]Block, 0)

			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Errorf(err.Error())
			}

			mutex.Lock()
			if len(chain) > len(BlockChain) {
				BlockChain = chain
				bytes, err := json.MarshalIndent(BlockChain, "", " ")
				if err != nil {
					log.Errorf(err.Error())
				}

				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			}
			mutex.Unlock()
		}
	}
}


func writeData(rw *bufio.ReadWriter) {

	go func() {
		for {
			time.Sleep(2 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(BlockChain)
			if err != nil {
				log.Errorf(err.Error())
			}
			mutex.Unlock()

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()
		}
	}()

	stdReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Errorf(err.Error())
		}

		sendData = strings.Replace(sendData, "\n", "", -1)
		bpm, err := strconv.Atoi(sendData)
		if err != nil {
			log.Errorf(err.Error())
		}

		// pick选择block生产者
		address := PickWinner()
		log.Infof(" %s ******", address)
		lastBlock := BlockChain[len(BlockChain)-1]
		newBlock, err := GenerateBlock(lastBlock, bpm, address)
		if err != nil {
			log.Errorf(err.Error())
		}

		if IsBlockValid(newBlock, lastBlock) {
			mutex.Lock()
			BlockChain = append(BlockChain, newBlock)
			mutex.Unlock()
		}

		spew.Dump(BlockChain)

		bytes, err := json.Marshal(BlockChain)
		if err != nil {
			log.Errorf(err.Error())
		}
		mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
		mutex.Unlock()
	}
}


func Run(ctx *cli.Context) error {

	t := time.Now()
	genesisBlock := Block{}
	genesisBlock = Block{0, t.String(), 0, CaculateBlockHash(genesisBlock), "", ""}
	BlockChain = append(BlockChain, genesisBlock)


	port := ctx.Int("port")
	target := ctx.String("target")
	secio := ctx.Bool("secio")
	seed := ctx.Int64("seed")

	if port == 0 {
		log.Fatal("")
	}
	ha, err := MakeBasicHost(port, secio, seed)
	if err != nil {
		return err
	}

	if target == "" {
		log.Info("...")
		ha.SetStreamHandler("/p2p/1.0.0", HandleStream)
		select {}
	} else {
		ha.SetStreamHandler("/p2p/1.0.0", HandleStream)
		ipfsaddr, err := ma.NewMultiaddr(target)
		if err != nil {
			return err
		}
		pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
		if err != nil {
			return err
		}

		peerid, err := peer.IDB58Decode(pid)
		if err != nil {
			return err
		}

		targetPeerAddr, _ := ma.NewMultiaddr(
			fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
		targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

		ha.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)
		log.Info("Stream")

		
		s, err := ha.NewStream(context.Background(), peerid, "/p2p/1.0.0")
		if err != nil {
			return err
		}

		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

		go writeData(rw)
		go readData(rw)
		select {}
	}
	return nil
}


func SavePeer(name string) {
	vote := DefaultVote 
	f, err := os.OpenFile(FileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Errorf(err.Error())
	}
	defer f.Close()

	f.WriteString(name + ":" + strconv.Itoa(vote) + "\n")

}
