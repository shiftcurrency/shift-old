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
        discover.MustParseNode("enode://4c8635f108dae8a997697d9c22ddca36969e7f9bc57d9fc01102d7e7d9633231331ae7f7307aceb1aa19130b5bdd4afe397db616c76e7ffc1c69302ba0d09a39@45.32.182.61:53900"),
        // Paris
        discover.MustParseNode("enode://80d0ce5c992f8cc83cdbfd6d832b2dff2e82fee1f8b58762cd858eaacfcc99d5a8a837648bd28a2d508cc1da305c15cf4e531546034ed1a8ccd07ff51a71abd6@108.61.177.0:53900"),
        // Seattle
        discover.MustParseNode("enode://f019da062a635a4e9e89ec93edc7ca11c06fdfec0574f1cb001126a82dc6ffa6ca05f924a683934ff5d01fc5d4b0ac9507349a945c97121b2a355d39b1781cd7@104.238.157.156:53900"),

}

// TestNetBootNodes are the enode URLs of the P2P bootstrap nodes running on the
// Morden test network.
var TestNetBootNodes = []*discover.Node{
}
