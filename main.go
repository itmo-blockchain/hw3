package main

import (
	"context"
	"flag"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	wnb = flag.Bool("without-new-block", false, "delete new block events")
	wnp = flag.Bool("without-new-price", false, "delete new price events")
)

type Config struct {
	Contracts []struct {
		Name    string `json:"name" yaml:"name"`
		Address string `json:"address" yaml:"address"`
	} `json:"contracts" yaml:"contracts"`
	NodeRPC string `json:"node_rpc" yaml:"nodeRPC"`
}

func main() {
	flag.Parse()
	//parse config from yaml file
	file, err := os.Open("config.yml")
	if err != nil {
		log.Fatal(err)
	}

	config := Config{}
	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	if config.NodeRPC == "" {
		config.NodeRPC = os.Getenv("ETHEREUM_RPC")
		if config.NodeRPC == "" {
			log.Fatal("no ethereum rpc url provided")
		}
	}

	client, err := ethclient.Dial(config.NodeRPC)
	if err != nil {
		log.Fatal(err)
	}

	contracts := make([]*ProxyWrapper, 0, len(config.Contracts))
	for _, contract := range config.Contracts {
		contract, err := NewContractClient(contract.Name, common.HexToAddress(contract.Address), client)
		if err != nil {
			log.Fatal(err)
		}
		contracts = append(contracts, contract)
	}

	sign := make(chan os.Signal, 1)
	signal.Notify(sign, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	if !*wnp {
		RunContractsListening(ctx, contracts)
	}

	if !*wnb {
		RunBlockListening(ctx, client)
	}

	<-sign
	cancel()
}

func RunBlockListening(ctx context.Context, client *ethclient.Client) {
	headers := make(chan *types.Header)

	head, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}

	go func(ctx context.Context, headers chan *types.Header) {
		for {
			select {
			case header := <-headers:
				log.Printf("New block number: %s\n", header.Number.String())
			case err := <-head.Err():
				log.Printf("subscription error: %v\n", err)
				return
			case <-ctx.Done():
				head.Unsubscribe()
				return
			}
		}
	}(ctx, headers)
}

func RunContractsListening(ctx context.Context, contracts []*ProxyWrapper) {
	for _, contract := range contracts {
		sub, err := contract.SubscribeOnUpdate()
		if err != nil {
			log.Fatal(err)
		}

		go func(ctx context.Context, contract *ProxyWrapper) {
			for {
				select {
				case ev := <-sub.data:
					price, err := contract.CalculatePrice(ev.Current)
					if err != nil {
						log.Printf("error while calculating price for contract %s: %v", contract.Name, err)
						continue
					}
					log.Println("New price for contract", contract.Name, ":", price)
				case err := <-sub.Err():
					log.Printf("subscription error: %v\n", err)
					return
				case <-ctx.Done():
					log.Println("context done")
					return
				}
			}
		}(ctx, contract)
	}
}
