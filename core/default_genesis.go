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
	"compress/gzip"
	"encoding/base64"
	"io"
	"strings"
)

func NewDefaultGenesisReader() (io.Reader, error) {
	return gzip.NewReader(base64.NewDecoder(base64.StdEncoding, strings.NewReader(defaultGenesisBlock)))
}

const defaultGenesisBlock = "H4sICMG68lYAA2dlbi50eHQArY9BDoIwFETv8tcsWoRiOYELLzEtrTRpC7E1wRDuLqIbE2Ji4ltOZv6fmSkOURtqiU3sgxJUUOesdfrm831z8JqxVYX3g6Z2JgHdoOpMVcoG9gjDuBRcHCC41tYKaVFxrsTTq+Dx+lTWbBdaloKCm3qkfqfQ76xV9eCiQtob+DWXXTApI4zv4CqNuJqYT38sd0E6u+Dydk9u4vIA48gkcZEBAAA="

