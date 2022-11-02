# Homework 3. Monitoring Ethereum blockchain

## Usage

### Build

For running the project you need docker installed on your machine. To build the project run:

```bash
docker build -t blockchain-monitor .
```

### Run

To run the project run with environment variable `ETHEREUM_RPC` set to the RPC endpoint of the Ethereum node you want to monitor:

```bash
docker run blockchain-monitor --env-file=".env"
```

You can pass the following flags to the program:

- -without-new-block - disables monitoring of new blocks
- -without-new-price - disables monitoring of new prices

## Example

```shell
2022/11/02 19:46:24 New block number: 15884514
2022/11/02 19:46:36 New block number: 15884515
2022/11/02 19:46:48 New block number: 15884516
...
2022/11/02 19:55:36 New block number: 15884560
2022/11/02 19:55:36 New price for contract EthUsdt : 1535.80504691
2022/11/02 19:55:48 New block number: 15884561
...
2022/11/02 20:11:50 New block number: 15884641
2022/11/02 20:11:50 New price for contract UsdtEth : 0.00065332973
2022/11/02 20:12:01 New block number: 15884642
...
2022/11/02 20:23:25 New block number: 15884699
2022/11/02 20:23:25 New price for contract LinkEth : 0.004898616614321285
2022/11/02 20:23:36 New block number: 15884700
...
```