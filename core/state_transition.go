// Copyright 2014 The go-ethereum Authors && Copyright 2015 shift Authors
// This file is part of the shift library.
//
// The shift library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The shift library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the shift library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"fmt"
	"math/big"

	"github.com/shiftcurrency/shift/common"
	"github.com/shiftcurrency/shift/core/vm"
	"github.com/shiftcurrency/shift/logger"
	"github.com/shiftcurrency/shift/logger/glog"
	"github.com/shiftcurrency/shift/params"

)

var (
	Big0 = big.NewInt(0)
)

/*
The State Transitioning Model

A state transition is a change made when a transaction is applied to the current world state
The state transitioning model does all all the necessary work to work out a valid new state root.

1) Nonce handling
2) Pre pay nrg
3) Create a new state object if the recipient is \0*32
4) Value transfer
== If contract creation ==
  4a) Attempt to run transaction data
  4b) If valid, use result as code for the new state object
== end ==
5) Run Script section
6) Derive new state root
*/
type StateTransition struct {
	gp            *NrgPool
	msg           Message
	nrg, nrgPrice *big.Int
	initialNrg    *big.Int
	value         *big.Int
	data          []byte
	state         vm.Database

	env vm.Environment
}

// Message represents a message sent to a contract.
type Message interface {
	From() (common.Address, error)
	FromFrontier() (common.Address, error)
	To() *common.Address

	NrgPrice() *big.Int
	Nrg() *big.Int
	Value() *big.Int

	Nonce() uint64
	Data() []byte
}

func MessageCreatesContract(msg Message) bool {
	return msg.To() == nil
}

// IntrinsicNrg computes the 'intrisic nrg' for a message
// with the given data.
func IntrinsicNrg(data []byte, contractCreation, homestead bool) *big.Int {
	inrg := new(big.Int)
	if contractCreation && homestead {
		inrg.Set(params.TxNrgContractCreation)
	} else {
		inrg.Set(params.TxNrg)
	}
	if len(data) > 0 {
		var nz int64
		for _, byt := range data {
			if byt != 0 {
				nz++
			}
		}
		m := big.NewInt(nz)
		m.Mul(m, params.TxDataNonZeroNrg)
		inrg.Add(inrg, m)
		m.SetInt64(int64(len(data)) - nz)
		m.Mul(m, params.TxDataZeroNrg)
		inrg.Add(inrg, m)
	}
	return inrg
}

func ApplyMessage(env vm.Environment, msg Message, gp *NrgPool) ([]byte, *big.Int, error) {
	var st = StateTransition{
		gp:         gp,
		env:        env,
		msg:        msg,
		nrg:        new(big.Int),
		nrgPrice:   msg.NrgPrice(),
		initialNrg: new(big.Int),
		value:      msg.Value(),
		data:       msg.Data(),
		state:      env.Db(),
	}
	return st.transitionDb()
}

func (self *StateTransition) from() (vm.Account, error) {
	var (
		f   common.Address
		err error
	)
	if params.IsHomestead(self.env.BlockNumber()) {
		f, err = self.msg.From()
	} else {
		f, err = self.msg.FromFrontier()
	}
	if err != nil {
		return nil, err
	}
	if !self.state.Exist(f) {
		return self.state.CreateAccount(f), nil
	}
	return self.state.GetAccount(f), nil
}
func (self *StateTransition) to() vm.Account {
	if self.msg == nil {
		return nil
	}
	to := self.msg.To()
	if to == nil {
		return nil // contract creation
	}

	if !self.state.Exist(*to) {
		return self.state.CreateAccount(*to)
	}
	return self.state.GetAccount(*to)
}

func (self *StateTransition) useNrg(amount *big.Int) error {
	if self.nrg.Cmp(amount) < 0 {
		return vm.OutOfNrgError
	}
	self.nrg.Sub(self.nrg, amount)

	return nil
}

func (self *StateTransition) addNrg(amount *big.Int) {
	self.nrg.Add(self.nrg, amount)
}

func (self *StateTransition) buyNrg() error {
	mnrg := self.msg.Nrg()
	mgval := new(big.Int).Mul(mnrg, self.nrgPrice)

	sender, err := self.from()
	if err != nil {
		return err
	}
	if sender.Balance().Cmp(mgval) < 0 {
		return fmt.Errorf("insufficient EXP for nrg (%x). Req %v, has %v", sender.Address().Bytes()[:4], mgval, sender.Balance())
	}
	if err = self.gp.SubNrg(mnrg); err != nil {
		return err
	}
	self.addNrg(mnrg)
	self.initialNrg.Set(mnrg)
	sender.SubBalance(mgval)
	return nil
}

func (self *StateTransition) preCheck() (err error) {
	msg := self.msg
	sender, err := self.from()
	if err != nil {
		return err
	}

	// Make sure this transaction's nonce is correct
	//if sender.Nonce() != msg.Nonce() {
	if n := self.state.GetNonce(sender.Address()); n != msg.Nonce() {
		return NonceError(msg.Nonce(), n)
	}

	// Pre-pay nrg
	if err = self.buyNrg(); err != nil {
		if IsNrgLimitErr(err) {
			return err
		}
		return InvalidTxError(err)
	}

	return nil
}

func (self *StateTransition) transitionDb() (ret []byte, usedNrg *big.Int, err error) {
	if err = self.preCheck(); err != nil {
		return
	}
	msg := self.msg
	sender, _ := self.from() // err checked in preCheck

	homestead := params.IsHomestead(self.env.BlockNumber())
	contractCreation := MessageCreatesContract(msg)
	// Pay intrinsic nrg
	if err = self.useNrg(IntrinsicNrg(self.data, contractCreation, homestead)); err != nil {
		return nil, nil, InvalidTxError(err)
	}

	vmenv := self.env
	//var addr common.Address
	if contractCreation {
		ret, _, err = vmenv.Create(sender, self.data, self.nrg, self.nrgPrice, self.value)
		if homestead && err == vm.CodeStoreOutOfNrgError {
			self.nrg = Big0
		}

		if err != nil {
			ret = nil
			glog.V(logger.Core).Infoln("VM create err:", err)
		}
	} else {
		// Increment the nonce for the next transaction
		self.state.SetNonce(sender.Address(), self.state.GetNonce(sender.Address())+1)
		ret, err = vmenv.Call(sender, self.to().Address(), self.data, self.nrg, self.nrgPrice, self.value)
		if err != nil {
			glog.V(logger.Core).Infoln("VM call err:", err)
		}
	}

	if err != nil && IsValueTransferErr(err) {
		return nil, nil, InvalidTxError(err)
	}

	// We aren't interested in errors here. Errors returned by the VM are non-consensus errors and therefor shouldn't bubble up
	if err != nil {
		err = nil
	}

	if vm.Debug {
		vm.StdErrFormat(vmenv.StructLogs())
	}

	self.refundNrg()
	self.state.AddBalance(self.env.Coinbase(), new(big.Int).Mul(self.nrgUsed(), self.nrgPrice))

	return ret, self.nrgUsed(), err
}

func (self *StateTransition) refundNrg() {
	// Return eth for remaining nrg to the sender account,
	// exchanged at the original rate.
	sender, _ := self.from() // err already checked
	remaining := new(big.Int).Mul(self.nrg, self.nrgPrice)
	sender.AddBalance(remaining)

	// Apply refund counter, capped to half of the used nrg.
	uhalf := remaining.Div(self.nrgUsed(), common.Big2)
	refund := common.BigMin(uhalf, self.state.GetRefund())
	self.nrg.Add(self.nrg, refund)
	self.state.AddBalance(sender.Address(), refund.Mul(refund, self.nrgPrice))

	// Also return remaining nrg to the block nrg counter so it is
	// available for the next transaction.
	self.gp.AddNrg(self.nrg)
}

func (self *StateTransition) nrgUsed() *big.Int {
	return new(big.Int).Sub(self.initialNrg, self.nrg)
}
