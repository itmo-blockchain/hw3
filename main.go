package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/ethclient"
)

type Config struct {
	Contracts struct {
		EthUsdt string `json:"eth_usdt" yaml:"EthUsdt"`
		LinkEth string `json:"link_eth" yaml:"LinkEth"`
		UsdtEth string `json:"usdt_eth" yaml:"UsdtEth"`
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

	ethUsdt, err := NewContractClient("ETH/USDT", common.HexToAddress(config.Contracts.EthUsdt), client)
	if err != nil {
		log.Fatal(err)
	}

	linkEth, err := NewContractClient("LINK/ETH", common.HexToAddress(config.Contracts.LinkEth), client)
	if err != nil {
		log.Fatal(err)
	}

	usdtEth, err := NewContractClient("USDT/ETH", common.HexToAddress(config.Contracts.UsdtEth), client)
	if err != nil {
		log.Fatal(err)
	}
	contracts := []*ContractClient{
		ethUsdt,
		linkEth,
		usdtEth,
	}

	sign := make(chan os.Signal, 1)
	signal.Notify(sign, os.Interrupt)

	for {
		select {
		case header := <-headers:
			fmt.Println(header.Number.String())

			for _, contract := range contracts {
				price, err := contract.GetPrice()
				if err != nil {
					log.Fatal(err)
				}

				time, err := contract.GetLatestTimestamp()
				if err != nil {
					log.Fatal(err)
				}

				fmt.Printf("Last update for %s: %d\n", contract.Name, time)
				fmt.Println(contract.Name, price)
			}

		case <-sign:
			head.Unsubscribe()
			break
		}
	}
}
