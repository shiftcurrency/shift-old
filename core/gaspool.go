// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import "math/big"

// NrgPool tracks the amount of nrg available during
// execution of the transactions in a block.
// The zero value is a pool with zero nrg available.
type NrgPool big.Int

// AddNrg makes nrg available for execution.
func (gp *NrgPool) AddNrg(amount *big.Int) *NrgPool {
	i := (*big.Int)(gp)
	i.Add(i, amount)
	return gp
}

// SubNrg deducts the given amount from the pool if enough nrg is
// available and returns an error otherwise.
func (gp *NrgPool) SubNrg(amount *big.Int) error {
	i := (*big.Int)(gp)
	if i.Cmp(amount) < 0 {
		return &NrgLimitErr{Have: new(big.Int).Set(i), Want: amount}
	}
	i.Sub(i, amount)
	return nil
}

func (gp *NrgPool) String() string {
	return (*big.Int)(gp).String()
}
