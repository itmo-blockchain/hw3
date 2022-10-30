package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	aggregator_v3_interface "hw3/contracts"
	"log"
	"math/big"
)

type ContractClient struct {
	Name     string
	contract *aggregator_v3_interface.AggregatorV3Interface
	address  common.Address
}

func NewContractClient(name string, address common.Address, client *ethclient.Client) (*ContractClient, error) {
	contract, err := aggregator_v3_interface.NewAggregatorV3Interface(address, client)
	if err != nil {
		return nil, err
	}
	return &ContractClient{
		Name:     name,
		contract: contract,
		address:  address,
	}, nil
}

func (c *ContractClient) GetLatestTimestamp() (int64, error) {
	latestRound, err := c.contract.LatestRoundData(nil)
	if err != nil {
		return 0, err
	}
	return latestRound.StartedAt.Int64(), nil
}

func (c *ContractClient) GetPrice() (*big.Float, error) {
	roundData, err := c.contract.LatestRoundData(nil)
	if err != nil {
		return nil, err
	}

	decimal, err := c.contract.Decimals(nil)
	if err != nil {
		return nil, err
	}

	decimalBigInt := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimal)), nil)
	price := divideBigInt(roundData.Answer, decimalBigInt)

	return price, nil
}

func divideBigInt(num1 *big.Int, num2 *big.Int) *big.Float {
	if num2.BitLen() == 0 {
		log.Fatal("cannot divide by zero.")
	}
	num1BigFloat := new(big.Float).SetInt(num1)
	num2BigFloat := new(big.Float).SetInt(num2)
	result := new(big.Float).Quo(num1BigFloat, num2BigFloat)
	return result
}
