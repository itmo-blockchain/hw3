package main

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/ethclient"
)

type Config struct {
	Contracts []struct {
		Name    string `json:"name" yaml:"name"`
		Address string `json:"address" yaml:"address"`
	} `json:"contracts" yaml:"contracts"`
	NodeRPC string `json:"node_rpc" yaml:"nodeRPC"`
}

func main() {
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

	client, err := ethclient.Dial(config.NodeRPC)
	if err != nil {
		log.Fatal(err)
	}

	headers := make(chan *types.Header)

	head, err := client.SubscribeNewHead(context.Background(), headers)
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

	for _, contract := range contracts {
		sub, err := contract.SubscribeOnUpdate()
		if err != nil {
			log.Fatal(err)
		}

		go func(contract *ProxyWrapper) {
			for {
				select {
				case ev := <-sub.data:
					log.Println("New data for contract", contract.Name, ":", ev)
				case err := <-sub.Err():
					log.Printf("subscription error: %v\n", err)
					return
				}
			}
		}(contract)
	}

	for {
		select {
		case header := <-headers:
			log.Printf("New block number: %s\n", header.Number.String())
		case <-sign:
			head.Unsubscribe()
			break
		}
	}
}
