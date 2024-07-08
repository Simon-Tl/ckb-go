package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/nervosnetwork/ckb-sdk-go/v2/collector"
	"github.com/nervosnetwork/ckb-sdk-go/v2/collector/builder"
	"github.com/nervosnetwork/ckb-sdk-go/v2/rpc"
	"github.com/nervosnetwork/ckb-sdk-go/v2/transaction/signer"
	"github.com/nervosnetwork/ckb-sdk-go/v2/types"
)

func SendChainTransaction() error {
	address1 := "ckt1qzda0cr08m85hc8jlnfp3zer7xulejywt49kt2rr"
	address2 := "ckt1qzda0cr08m85hc8jlnfp3zer7xulejywt49kt2rr"
	address3 := "ckt1qzda0cr08m85hc8jlnfp3zer7xulejywt49kt2rr"

	network := types.NetworkTest
	client, err := rpc.Dial("https://testnet.ckb.dev")
	if err != nil {
		return err
	}
	offChainInputCollector := collector.NewOffChainInputCollector(client)

	var iterator, err_ = collector.NewOffChainInputIteratorFromAddress(client, address1, offChainInputCollector, true)
	if err_ != nil {
		return err
	}

	txWithGroupsBuilder := builder.NewCkbTransactionBuilder(network, iterator)
	err = txWithGroupsBuilder.AddOutputByAddress(address2, 501000000000)
	if err != nil {
		return err
	}

	err = txWithGroupsBuilder.AddChangeOutputByAddress(address1)
	if err != nil {
		return err
	}

	config := new(signer.TransactionSigner)
	txWithGroups, err := txWithGroupsBuilder.Build(config)

	if err != nil {
		return err
	}

	// sign transaction
	if _, err = signer.GetTransactionSignerInstance(network).SignTransactionByPrivateKeys(txWithGroups, "0xxx"); err != nil {
		return err
	}

	// send transaction
	hash, err := client.SendTransaction(context.Background(), txWithGroups.TxView)

	if err != nil {
		return err
	}

	blockNumber, err := client.GetTipBlockNumber(context.Background())

	if err != nil {
		return err
	}

	offChainInputCollector.ApplyOffChainTransaction(blockNumber, *txWithGroups.TxView)
	fmt.Println("transaction hash: " + hexutil.Encode(hash.Bytes()))

	it2, _ := collector.NewLiveCellIteratorFromAddress(client, address2)
	cellIterator2 := it2.(*collector.LiveCellIterator)
	var iterator2 = collector.OffChainInputIterator{
		Iterator:                    cellIterator2,
		Collector:                   offChainInputCollector,
		ConsumeOffChainCellsFirstly: true,
	}

	txWithGroupsBuilder = builder.NewCkbTransactionBuilder(network, &iterator2)
	err = txWithGroupsBuilder.AddOutputByAddress(address3, 100000000000)
	if err != nil {
		return err
	}

	if err = txWithGroupsBuilder.AddChangeOutputByAddress(address2); err != nil {
		return err
	}

	txWithGroups, err = txWithGroupsBuilder.Build(config)

	// sign transaction
	txSigner := signer.GetTransactionSignerInstance(network)
	_, err = txSigner.SignTransactionByPrivateKeys(txWithGroups, "0xxxx")
	if err != nil {
		return err
	}

	// send transaction
	hash, err = client.SendTransaction(context.Background(), txWithGroups.TxView)

	if err != nil {
		return err
	}

	blockNumber, err = client.GetTipBlockNumber(context.Background())

	if err != nil {
		return err
	}

	offChainInputCollector.ApplyOffChainTransaction(blockNumber, *txWithGroups.TxView)
	fmt.Println("transaction hash: " + hexutil.Encode(hash.Bytes()))

	return nil
}

func main() {
	err := SendChainTransaction()
	if err != nil {
		fmt.Println(err)
	}
}
