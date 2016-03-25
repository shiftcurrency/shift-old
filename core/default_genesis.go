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

const defaultGenesisBlock = "H4sICLcQ9VYAA2dlbi50eHQArY/BCsIwEET/Zc89JLGmpl/gwZ/Yxo1dSNJiIlRK/91YvQhFEJzjY2Z3ZoY4REvQgpjEhxRCBWd2ju3N5/vqUIUXit4PFtoZNFKDNVGtTIPuQCSk0VLvUEtrndPGYS1lp5/eDj2+Pqm92BQsSwWBpx5Tv1Hod5WqduDYYdoa+DWXOVDKGMZ3sKARrxTz8Y/lLphOHDiv98wKlwdI1OM1kQEAAA=="
