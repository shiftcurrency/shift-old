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

package vm

import (
	"math/big"

	"github.com/shiftcurrency/shift/common"
)

// ContractRef is a reference to the contract's backing object
type ContractRef interface {
	ReturnNrg(*big.Int, *big.Int)
	Address() common.Address
	Value() *big.Int
	SetCode([]byte)
}

// Contract represents an ethereum contract in the state database. It contains
// the the contract code, calling arguments. Contract implements ContractRef
type Contract struct {
	// CallerAddress is the result of the caller which initialised this
	// contract. However when the "call method" is delegated this value
	// needs to be initialised to that of the caller's caller.
	CallerAddress common.Address
	caller        ContractRef
	self          ContractRef

	jumpdests destinations // result of JUMPDEST analysis.

	Code     []byte
	Input    []byte
	CodeAddr *common.Address

	value, Nrg, UsedNrg, Price *big.Int

	Args []byte

	DelegateCall bool
}

// NewContract returns a new contract environment for the execution of EVM.
func NewContract(caller ContractRef, object ContractRef, value, nrg, price *big.Int) *Contract {
	c := &Contract{CallerAddress: caller.Address(), caller: caller, self: object, Args: nil}

	if parent, ok := caller.(*Contract); ok {
		// Reuse JUMPDEST analysis from parent context if available.
		c.jumpdests = parent.jumpdests
	} else {
		c.jumpdests = make(destinations)
	}

	// Nrg should be a pointer so it can safely be reduced through the run
	// This pointer will be off the state transition
	c.Nrg = nrg //new(big.Int).Set(nrg)
	c.value = new(big.Int).Set(value)
	// In most cases price and value are pointers to transaction objects
	// and we don't want the transaction's values to change.
	c.Price = new(big.Int).Set(price)
	c.UsedNrg = new(big.Int)

	return c
}

// AsDelegate sets the contract to be a delegate call and returns the current
// contract (for chaining calls)
func (c *Contract) AsDelegate() *Contract {
	c.DelegateCall = true
	// NOTE: caller must, at all times be a contract. It should never happen
	// that caller is something other than a Contract.
	c.CallerAddress = c.caller.(*Contract).CallerAddress
	return c
}

// GetOp returns the n'th element in the contract's byte array
func (c *Contract) GetOp(n uint64) OpCode {
	return OpCode(c.GetByte(n))
}

// GetByte returns the n'th byte in the contract's byte array
func (c *Contract) GetByte(n uint64) byte {
	if n < uint64(len(c.Code)) {
		return c.Code[n]
	}

	return 0
}

// Caller returns the caller of the contract.
//
// Caller will recursively call caller when the contract is a delegate
// call, including that of caller's caller.
func (c *Contract) Caller() common.Address {
	return c.CallerAddress
}

// Finalise finalises the contract and returning any remaining nrg to the original
// caller.
func (c *Contract) Finalise() {
	// Return the remaining nrg to the caller
	c.caller.ReturnNrg(c.Nrg, c.Price)
}

// UseNrg attempts the use nrg and subtracts it and returns true on success
func (c *Contract) UseNrg(nrg *big.Int) (ok bool) {
	ok = useNrg(c.Nrg, nrg)
	if ok {
		c.UsedNrg.Add(c.UsedNrg, nrg)
	}
	return
}

// ReturnNrg adds the given nrg back to itself.
func (c *Contract) ReturnNrg(nrg, price *big.Int) {
	// Return the nrg to the context
	c.Nrg.Add(c.Nrg, nrg)
	c.UsedNrg.Sub(c.UsedNrg, nrg)
}

// Address returns the contracts address
func (c *Contract) Address() common.Address {
	return c.self.Address()
}

// Value returns the contracts value (sent to it from it's caller)
func (c *Contract) Value() *big.Int {
	return c.value
}

// SetCode sets the code to the contract
func (self *Contract) SetCode(code []byte) {
	self.Code = code
}

// SetCallCode sets the code of the contract and address of the backing data
// object
func (self *Contract) SetCallCode(addr *common.Address, code []byte) {
	self.Code = code
	self.CodeAddr = addr
}
