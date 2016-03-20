// Copyright 2015 The shift Authors
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
	"fmt"
	"math/big"

	"github.com/shiftcurrency/shift/params"
)

var (
	NrgQuickStep   = big.NewInt(2)
	NrgFastestStep = big.NewInt(3)
	NrgFastStep    = big.NewInt(5)
	NrgMidStep     = big.NewInt(8)
	NrgSlowStep    = big.NewInt(10)
	NrgExtStep     = big.NewInt(20)

	NrgReturn = big.NewInt(0)
	NrgStop   = big.NewInt(0)

	NrgContractByte = big.NewInt(200)
)

// baseCheck checks for any stack error underflows
func baseCheck(op OpCode, stack *stack, nrg *big.Int) error {
	// PUSH and DUP are a bit special. They all cost the same but we do want to have checking on stack push limit
	// PUSH is also allowed to calculate the same price for all PUSHes
	// DUP requirements are handled elsewhere (except for the stack limit check)
	if op >= PUSH1 && op <= PUSH32 {
		op = PUSH1
	}
	if op >= DUP1 && op <= DUP16 {
		op = DUP1
	}

	if r, ok := _baseCheck[op]; ok {
		err := stack.require(r.stackPop)
		if err != nil {
			return err
		}

		if r.stackPush > 0 && stack.len()-r.stackPop+r.stackPush > int(params.StackLimit.Int64()) {
			return fmt.Errorf("stack limit reached %d (%d)", stack.len(), params.StackLimit.Int64())
		}

		nrg.Add(nrg, r.nrg)
	}
	return nil
}

// casts a arbitrary number to the amount of words (sets of 32 bytes)
func toWordSize(size *big.Int) *big.Int {
	tmp := new(big.Int)
	tmp.Add(size, u256(31))
	tmp.Div(tmp, u256(32))
	return tmp
}

type req struct {
	stackPop  int
	nrg       *big.Int
	stackPush int
}

var _baseCheck = map[OpCode]req{
	// opcode  |  stack pop | nrg price | stack push
	ADD:          {2, NrgFastestStep, 1},
	LT:           {2, NrgFastestStep, 1},
	GT:           {2, NrgFastestStep, 1},
	SLT:          {2, NrgFastestStep, 1},
	SGT:          {2, NrgFastestStep, 1},
	EQ:           {2, NrgFastestStep, 1},
	ISZERO:       {1, NrgFastestStep, 1},
	SUB:          {2, NrgFastestStep, 1},
	AND:          {2, NrgFastestStep, 1},
	OR:           {2, NrgFastestStep, 1},
	XOR:          {2, NrgFastestStep, 1},
	NOT:          {1, NrgFastestStep, 1},
	BYTE:         {2, NrgFastestStep, 1},
	CALLDATALOAD: {1, NrgFastestStep, 1},
	CALLDATACOPY: {3, NrgFastestStep, 1},
	MLOAD:        {1, NrgFastestStep, 1},
	MSTORE:       {2, NrgFastestStep, 0},
	MSTORE8:      {2, NrgFastestStep, 0},
	CODECOPY:     {3, NrgFastestStep, 0},
	MUL:          {2, NrgFastStep, 1},
	DIV:          {2, NrgFastStep, 1},
	SDIV:         {2, NrgFastStep, 1},
	MOD:          {2, NrgFastStep, 1},
	SMOD:         {2, NrgFastStep, 1},
	SIGNEXTEND:   {2, NrgFastStep, 1},
	ADDMOD:       {3, NrgMidStep, 1},
	MULMOD:       {3, NrgMidStep, 1},
	JUMP:         {1, NrgMidStep, 0},
	JUMPI:        {2, NrgSlowStep, 0},
	EXP:          {2, NrgSlowStep, 1},
	ADDRESS:      {0, NrgQuickStep, 1},
	ORIGIN:       {0, NrgQuickStep, 1},
	CALLER:       {0, NrgQuickStep, 1},
	CALLVALUE:    {0, NrgQuickStep, 1},
	CODESIZE:     {0, NrgQuickStep, 1},
	GASPRICE:     {0, NrgQuickStep, 1},
	COINBASE:     {0, NrgQuickStep, 1},
	TIMESTAMP:    {0, NrgQuickStep, 1},
	NUMBER:       {0, NrgQuickStep, 1},
	CALLDATASIZE: {0, NrgQuickStep, 1},
	DIFFICULTY:   {0, NrgQuickStep, 1},
	GASLIMIT:     {0, NrgQuickStep, 1},
	POP:          {1, NrgQuickStep, 0},
	PC:           {0, NrgQuickStep, 1},
	MSIZE:        {0, NrgQuickStep, 1},
	GAS:          {0, NrgQuickStep, 1},
	BLOCKHASH:    {1, NrgExtStep, 1},
	BALANCE:      {1, NrgExtStep, 1},
	EXTCODESIZE:  {1, NrgExtStep, 1},
	EXTCODECOPY:  {4, NrgExtStep, 0},
	SLOAD:        {1, params.SloadNrg, 1},
	SSTORE:       {2, Zero, 0},
	SHA3:         {2, params.Sha3Nrg, 1},
	CREATE:       {3, params.CreateNrg, 1},
	CALL:         {7, params.CallNrg, 1},
	CALLCODE:     {7, params.CallNrg, 1},
	DELEGATECALL: {6, params.CallNrg, 1},
	JUMPDEST:     {0, params.JumpdestNrg, 0},
	SUICIDE:      {1, Zero, 0},
	RETURN:       {2, Zero, 0},
	PUSH1:        {0, NrgFastestStep, 1},
	DUP1:         {0, Zero, 1},
}
