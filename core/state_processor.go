package core

import (
	"math/big"

	"github.com/shiftcurrency/shift/core/state"
	"github.com/shiftcurrency/shift/core/types"
	"github.com/shiftcurrency/shift/core/vm"
	"github.com/shiftcurrency/shift/crypto"
	"github.com/shiftcurrency/shift/logger"
	"github.com/shiftcurrency/shift/logger/glog"
)

var (
	big8  = big.NewInt(8)
	big32 = big.NewInt(32)
)

type StateProcessor struct {
	bc *BlockChain
}

func NewStateProcessor(bc *BlockChain) *StateProcessor {
	return &StateProcessor{bc}
}

// Process processes the state changes according to the Ethereum rules by running
// the transaction messages using the statedb and applying any rewards to both
// the processor (shiftbase) and any included uncles.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) Process(block *types.Block, statedb *state.StateDB) (types.Receipts, vm.Logs, *big.Int, error) {
	var (
		receipts     types.Receipts
		totalUsedGas = big.NewInt(0)
		err          error
		header       = block.Header()
		allLogs      vm.Logs
		gp           = new(GasPool).AddGas(block.GasLimit())
	)

	for i, tx := range block.Transactions() {
		statedb.StartRecord(tx.Hash(), block.Hash(), i)
		receipt, logs, _, err := ApplyTransaction(p.bc, gp, statedb, header, tx, totalUsedGas)
		if err != nil {
			return nil, nil, totalUsedGas, err
		}
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, logs...)
	}
	AccumulateRewards(statedb, header, block.Uncles(), block)

	return receipts, allLogs, totalUsedGas, err
}

// ApplyTransaction attemps to apply a transaction to the given state database
// and uses the input parameters for its environment.
//
// ApplyTransactions returns the generated receipts and vm logs during the
// execution of the state transition phase.
func ApplyTransaction(bc *BlockChain, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *big.Int) (*types.Receipt, vm.Logs, *big.Int, error) {
	_, gas, err := ApplyMessage(NewEnv(statedb, bc, tx, header), tx, gp)
	if err != nil {
		return nil, nil, nil, err
	}

	// Update the state with pending changes
	usedGas.Add(usedGas, gas)
	receipt := types.NewReceipt(statedb.IntermediateRoot().Bytes(), usedGas)
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = new(big.Int).Set(gas)
	if MessageCreatesContract(tx) {
		from, _ := tx.From()
		receipt.ContractAddress = crypto.CreateAddress(from, tx.Nonce())
	}

	logs := statedb.GetLogs(tx.Hash())
	receipt.Logs = logs
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	glog.V(logger.Debug).Infoln(receipt)

	return receipt, logs, gas, err
}

// AccumulateRewards credits the shiftbase of the given block with the
// mining reward. The total reward consists of the static block reward
// and rewards for included uncles. The shiftbase of each uncle block is
// also rewarded.
func AccumulateRewards(statedb *state.StateDB, header *types.Header, uncles []*types.Header, block *types.Block) {

    reward := new(big.Int).Set(BlockReward)

    // FIXME: INVALID MERKLE ROOT BECAUSE OF...

    // 80 days decay of mining reward. From 3 to 1 SHIFT.
    /*
    if blockNum >= 28800 && blockNum < 57600 {
        reward = new(big.Int).Set(BRD2)
    } else if blockNum >= 57600 && blockNum < 86400 {
        reward = new(big.Int).Set(BRD3)
    } else if blockNum >= 86400 && blockNum < 115200 {
        reward = new(big.Int).Set(BRD4)
    } else if blockNum >= 115200 && blockNum < 144000 {
        reward = new(big.Int).Set(BRD5) 
    } else if blockNum >= 144000 && blockNum < 172800 {
        reward = new(big.Int).Set(BRD6)
    } else if blockNum >= 172800 && blockNum < 230400 {
        reward = new(big.Int).Set(BRD7)
    } else if blockNum >= 230400 {
        reward = new(big.Int).Set(BRD8)
    }*/

	r := new(big.Int)
	for _, uncle := range uncles {
		r.Add(uncle.Number, big8)
		r.Sub(r, header.Number)
		r.Mul(r, BlockReward)
		r.Div(r, big8)
		statedb.AddBalance(uncle.Coinbase, r)

		r.Div(BlockReward, big32)
		reward.Add(reward, r)
	}
	statedb.AddBalance(header.Coinbase, reward)
}
