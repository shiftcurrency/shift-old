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

package state

import (
	"fmt"
	"math/big"
)

type NrgLimitErr struct {
	Message string
	Is, Max *big.Int
}

func IsNrgLimitErr(err error) bool {
	_, ok := err.(*NrgLimitErr)

	return ok
}
func (err *NrgLimitErr) Error() string {
	return err.Message
}
func NrgLimitError(is, max *big.Int) *NrgLimitErr {
	return &NrgLimitErr{Message: fmt.Sprintf("NrgLimit error. Max %s, transaction would take it to %s", max, is), Is: is, Max: max}
}
