package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	proxy "hw3/contracts"
	aggregator "hw3/contracts/interfaces"
	"log"
	"math/big"
)

type ProxyWrapper struct {
	Name     string
	contract *proxy.ProxyContract
	address  common.Address
	cl       *ethclient.Client
}

func NewContractClient(name string, address common.Address, client *ethclient.Client) (*ProxyWrapper, error) {
	contract, err := proxy.NewProxyContract(address, client)
	if err != nil {
		return nil, err
	}
	return &ProxyWrapper{
		Name:     name,
		contract: contract,
		address:  address,
		cl:       client,
	}, nil
}

func (c *ProxyWrapper) GetLatestTimestamp() (int64, error) {
	latestRound, err := c.contract.LatestRoundData(nil)
	if err != nil {
		return 0, err
	}
	return latestRound.StartedAt.Int64(), nil
}

func (c *ProxyWrapper) GetPrice() (*big.Float, error) {
	roundData, err := c.contract.LatestRoundData(nil)
	if err != nil {
		return nil, err
	}

	return c.CalculatePrice(roundData.Answer)
}

func (c *ProxyWrapper) CalculatePrice(a *big.Int) (*big.Float, error) {
	decimal, err := c.contract.Decimals(nil)
	if err != nil {
		return nil, err
	}

	decimalBigInt := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimal)), nil)
	price := divideBigInt(a, decimalBigInt)

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

type Subscription struct {
	event.Subscription
	data <-chan *aggregator.AggregatorInterfaceAnswerUpdated
}

func (c *ProxyWrapper) SubscribeOnUpdate() (*Subscription, error) {

	aggregAddr, err := c.contract.Aggregator(nil)
	if err != nil {
		return nil, err
	}

	aggreg, err := aggregator.NewAggregatorInterface(aggregAddr, c.cl)

	answerUpdated := make(chan *aggregator.AggregatorInterfaceAnswerUpdated)
	updated, err := aggreg.WatchAnswerUpdated(nil, answerUpdated, nil, nil)
	if err != nil {
		return nil, err
	}

	return &Subscription{
		Subscription: updated,
		data:         answerUpdated,
	}, nil
}
