package transactions

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/nervosnetwork/ckb-sdk-go/v2/address"
	"github.com/nervosnetwork/ckb-sdk-go/v2/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/v2/rpc"
	"github.com/nervosnetwork/ckb-sdk-go/v2/systemscript"
	"github.com/nervosnetwork/ckb-sdk-go/v2/transaction"
	"github.com/nervosnetwork/ckb-sdk-go/v2/transaction/signer"
	"github.com/nervosnetwork/ckb-sdk-go/v2/types"
	"log"
	"math/big"
)

func GenericSudtCell() error {

	sudtCapacity := uint64(143000000000)
	client, err := rpc.Dial("https://testnet.ckb.dev")
	if err != nil {
		log.Fatal(err)
	}
	//addr, err := address.Decode("ckt1qzda0cr08m85hc8jlnfp3zer7xulejywt49kt2rr0vthywaa50xwsqt8z2mch5exaj8zdg258ws270g9pahlmhgnavc2h")
	// 准备 SUDT 类型脚本
	a := systemscript.GetInfo(types.NetworkTest, systemscript.Sudt)
	addr, err := address.Decode("ckt1qzda0cr08m85hc8jlnfp3zer7xulejywt49kt2rr0vthywaa5")
	if err != nil {
		log.Fatal(err)
	}

	searchKey := &indexer.SearchKey{
		Script: &types.Script{
			CodeHash: addr.Script.CodeHash,
			HashType: addr.Script.HashType,
			Args:     addr.Script.Args,
		},
		ScriptType: types.ScriptTypeLock,
	}

	liveCells, err := client.GetCells(context.Background(), searchKey, indexer.SearchOrderAsc, 100, "")
	if err != nil {
		log.Fatal(err)
	}

	var totolCapacity uint64
	var sender *indexer.LiveCell
	for _, cell := range liveCells.Objects {
		if cell.Output.Capacity > sudtCapacity {
			totolCapacity = cell.Output.Capacity
			sender = cell
			fmt.Println("capacity enough for", cell.Output.Lock.Hash())
			break

		}

	}

	//hashType := types.HashTypeType

	sudtType := &types.Script{
		CodeHash: a.CodeHash,
		HashType: a.HashType,
		Args:     sender.Output.Lock.Hash().Bytes(), // 通常是发行者的 Lock Script 的哈希
	}

	// 准备 SUDT Cell 的输出
	sudtCellOutput := &types.CellOutput{
		Capacity: sudtCapacity, // 200 CKB
		Lock: &types.Script{
			CodeHash: addr.Script.CodeHash, // 锁脚本的 code_hash
			HashType: addr.Script.HashType,
			Args:     addr.Script.Args, // 锁脚本的参数
		},
		Type: sudtType,
	}

	ckbCellOutput := &types.CellOutput{
		Capacity: totolCapacity - sudtCapacity,
		Lock: &types.Script{
			CodeHash: addr.Script.CodeHash, // 锁脚本的 code_hash
			HashType: addr.Script.HashType,
			Args:     addr.Script.Args, // 锁脚本的参数
		},
		Type: nil,
	}
	// 设置 SUDT Cell 的数据
	sudtAmount := big.NewInt(10000)
	sudtdata := systemscript.EncodeSudtAmount(sudtAmount)

	// 构建交易
	tx := &types.Transaction{
		Version: 0,
		CellDeps: []*types.CellDep{
			{
				OutPoint: &types.OutPoint{
					TxHash: types.HexToHash("0xf8de3bb47d055cdf460d93a2a6e1b05f7432f9777c8c474abf4eec1d4aee5d37"),
					Index:  0,
				},
				DepType: types.DepTypeDepGroup,
			},
			{
				OutPoint: &types.OutPoint{
					TxHash: types.HexToHash("0xe12877ebd2c3c364dc46c5c992bcfaf4fee33fa13eebdf82c591fc9825aab769"), // SUDT 合约所在的交易哈希
					Index:  0,
				},
				DepType: types.DepTypeCode,
			},
		},
		Inputs: []*types.CellInput{
			{
				PreviousOutput: &types.OutPoint{
					TxHash: sender.OutPoint.TxHash, // 输入 Cell 的交易哈希
					Index:  sender.OutPoint.Index,
				},
				Since: 0,
			},
		},
		Outputs:     []*types.CellOutput{sudtCellOutput, ckbCellOutput},
		OutputsData: [][]byte{sudtdata, {}},
		Witnesses:   [][]byte{{0, 0, 0, 0}},
	}
	//ckbCellOutput.Capacity -= sudtCapacity
	//tx.Outputs[0] = ckbCellOutput
	fee := tx.CalculateFee(1000)
	sudtCellOutput.Capacity -= fee
	ckbCellOutput.Capacity -= fee
	tx.Outputs[0] = sudtCellOutput
	tx.Outputs[1] = ckbCellOutput
	// 签名交易

	scriptGroups := []*transaction.ScriptGroup{
		{
			Script: &types.Script{
				CodeHash: addr.Script.CodeHash,
				HashType: addr.Script.HashType,
				Args:     addr.Script.Args,
			},
			GroupType:     types.ScriptTypeLock,
			InputIndices:  []uint32{0},
			OutputIndices: []uint32{1},
		},
	}
	txWithScriptGroups := &transaction.TransactionWithScriptGroups{
		TxView:       tx,
		ScriptGroups: scriptGroups,
	}

	// 构建 WitnessArgs
	witnessArgs := types.WitnessArgs{
		Lock:       make([]byte, 65), // Lock 是签名后的结果，长度为 65 字节
		InputType:  nil,
		OutputType: nil,
	}
	witnessArgsSerialized := witnessArgs.Serialize()
	if err != nil {
		log.Fatal(err)
	}

	// 更新 Witnesses 数据
	tx.Witnesses[0] = witnessArgsSerialized
	privKey := ""
	// 1. Sign transaction with your private key
	txSigner := signer.GetTransactionSignerInstance(types.NetworkTest)
	_, err = txSigner.SignTransactionByPrivateKeys(txWithScriptGroups, privKey)
	if err != nil {
		fmt.Println("签名错误", err)
	}

	// 发送交易
	txHash, err := client.SendTransaction(context.Background(), txWithScriptGroups.TxView)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Transaction hash: %s\n", txHash.String())
	return err
}

// QuerySudtAmount query sudt blancer
func QuerySudtAmount(add string) (err error, res *int, cellinfo []*indexer.LiveCell) {
	// 初始化 CKB 客户端
	client, err := rpc.Dial("https://testnet.ckb.dev")
	if err != nil {
		log.Fatal(err)
	}
	a := systemscript.GetInfo(types.NetworkTest, systemscript.Sudt)
	addr, err := address.Decode(add)
	if err != nil {
		log.Fatal(err)
	}

	searchKey := &indexer.SearchKey{
		Script: &types.Script{
			CodeHash: addr.Script.CodeHash,
			HashType: addr.Script.HashType,
			Args:     addr.Script.Args,
		},
		ScriptType: types.ScriptTypeLock,
	}

	liveCells, err := client.GetCells(context.Background(), searchKey, indexer.SearchOrderAsc, 100, "")
	if err != nil {
		log.Fatal(err)
	}
	result := big.NewInt(0)
	var sender []*indexer.LiveCell
	for _, cell := range liveCells.Objects {
		if cell.Output.Type != nil && cell.Output.Type.CodeHash == a.CodeHash {
			data, _ := client.GetLiveCell(context.Background(), cell.OutPoint, true)
			amount, _ := systemscript.DecodeSudtAmount(data.Cell.Data.Content)
			result.Add(result, amount)
			sender = append(sender, cell)

		}

	}
	fmt.Println(result)

	return err, res, sender
}

func TransactionSudt() error {
	sudtCapacity := uint64(14300000000)
	client, err := rpc.Dial("https://testnet.ckb.dev")
	if err != nil {
		log.Fatal(err)
	}

	send, err := address.Decode("")
	receive, err := address.Decode("")
	if err != nil {
		log.Fatal(err)
	}

	sendliveCells, err := client.GetCells(context.Background(), &indexer.SearchKey{
		Script: &types.Script{
			CodeHash: send.Script.CodeHash,
			HashType: send.Script.HashType,
			Args:     send.Script.Args,
		},
		ScriptType: types.ScriptTypeLock,
	}, indexer.SearchOrderAsc, 100, "")
	if err != nil {
		log.Fatal(err)
	}
	a := systemscript.GetInfo(types.NetworkTest, systemscript.Sudt)
	var totalCapacity uint64
	var inputCells []*types.CellInput
	for _, cell := range sendliveCells.Objects {
		if cell.Output.Type != nil && cell.Output.Type.CodeHash == a.CodeHash {
			totalCapacity += cell.Output.Capacity
			inputCells = append(inputCells, &types.CellInput{
				PreviousOutput: cell.OutPoint,
				Since:          0,
			})
		}
		if totalCapacity > sudtCapacity {
			break
		}
	}
	fmt.Println(totalCapacity, sudtCapacity)

	if totalCapacity < sudtCapacity {
		log.Fatal("not enough capacity")
	}

	amount := uint64(2000)

	// Prepare outputs
	transferOutput := &types.CellOutput{
		Capacity: sudtCapacity,
		Lock: &types.Script{
			CodeHash: receive.Script.CodeHash,
			HashType: receive.Script.HashType,
			Args:     receive.Script.Args,
		},
		Type: &types.Script{
			CodeHash: a.CodeHash,
			HashType: a.HashType,
			Args:     send.Script.Hash().Bytes(),
		},
	}
	changeOutput := &types.CellOutput{
		Capacity: totalCapacity - sudtCapacity,
		Lock: &types.Script{
			CodeHash: send.Script.CodeHash,
			HashType: send.Script.HashType,
			Args:     send.Script.Args,
		},
		Type: &types.Script{
			CodeHash: a.CodeHash,
			HashType: a.HashType,
			Args:     send.Script.Hash().Bytes(),
		},
	}

	// SUDT amounts
	transferAmount := make([]byte, 16)
	binary.LittleEndian.PutUint64(transferAmount, amount)
	changeAmount := make([]byte, 16)
	binary.LittleEndian.PutUint64(changeAmount, 10000-amount)

	// Construct transaction
	tx := &types.Transaction{
		Version: 0,
		CellDeps: []*types.CellDep{
			{
				OutPoint: &types.OutPoint{
					TxHash: types.HexToHash("0xf8de3bb47d055cdf460d93a2a6e1b05f7432f9777c8c474abf4eec1d4aee5d37"),
					Index:  0,
				},
				DepType: types.DepTypeDepGroup,
			},
			{
				OutPoint: &types.OutPoint{
					TxHash: types.HexToHash("0xe12877ebd2c3c364dc46c5c992bcfaf4fee33fa13eebdf82c591fc9825aab769"),
					Index:  0,
				},
				DepType: types.DepTypeCode,
			},
		},
		Inputs:      inputCells,
		Outputs:     []*types.CellOutput{transferOutput, changeOutput},
		OutputsData: [][]byte{transferAmount, changeAmount},
		Witnesses:   [][]byte{{}},
	}

	// Calculate fee and adjust change output
	fee := tx.CalculateFee(1000)
	transferOutput.Capacity -= fee
	changeOutput.Capacity -= fee
	tx.Outputs[0] = transferOutput
	tx.Outputs[1] = changeOutput

	// Construct script groups for signing
	scriptGroups := []*transaction.ScriptGroup{
		{
			Script:       send.Script,
			GroupType:    types.ScriptTypeLock,
			InputIndices: []uint32{0},
		},
	}

	// Sign transaction
	txWithScriptGroups := &transaction.TransactionWithScriptGroups{
		TxView:       tx,
		ScriptGroups: scriptGroups,
	}
	// 构建 WitnessArgs
	witnessArgs := types.WitnessArgs{
		Lock:       make([]byte, 65), // Lock 是签名后的结果，长度为 65 字节
		InputType:  nil,
		OutputType: nil,
	}
	witnessArgsSerialized := witnessArgs.Serialize()
	if err != nil {
		log.Fatal(err)
	}

	// 更新 Witnesses 数据
	tx.Witnesses[0] = witnessArgsSerialized

	// Sign transaction with private key
	privKey := ""
	txSigner := signer.GetTransactionSignerInstance(types.NetworkTest)
	_, err = txSigner.SignTransactionByPrivateKeys(txWithScriptGroups, privKey)
	if err != nil {
		fmt.Println("签名错误", err)
		return err
	}

	// Send transaction
	txHash, err := client.SendTransaction(context.Background(), txWithScriptGroups.TxView)
	if err != nil {
		log.Fatal("交易发送失败", err)
	}

	fmt.Printf("Transaction hash: %s\n", txHash.String())
	return nil
}
