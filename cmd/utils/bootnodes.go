// Copyright 2015 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package utils

import "github.com/shiftcurrency/shift/p2p/discover"

// FrontierBootNodes are the enode URLs of the P2P bootstrap nodes running on
// the Frontier network.
var FrontierBootNodes = []*discover.Node{
        discover.MustParseNode("enode://3ff38d26ff93db0c340acfae3999595c52c170fa8a360baca38533eaf6cbec4c7b9cd9f3da4a59bfa27dfa26ab856a5b8f3f3f0f91a61b98da6774e986800747@45.32.95.111:54900"),
        // Paris
        discover.MustParseNode("enode://7eff1768d648c0df27ca6b808af28a6ab5e2a30fda34a0c1c6ffb637c09f85499571329dcd13bd434912ba2f8e4af3a38eb51b5b0dcf8320e2d795b561214bf3@45.32.155.17:54900"),

}

// TestNetBootNodes are the enode URLs of the P2P bootstrap nodes running on the
// Morden test network.
var TestNetBootNodes = []*discover.Node{
}
