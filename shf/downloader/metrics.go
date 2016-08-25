// Copyright 2015 The go-ethereum Authors
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

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/shiftcurrency/shift/metrics"
)

var (
	headerInMeter      = metrics.NewMeter("shf/downloader/headers/in")
	headerReqTimer     = metrics.NewTimer("shf/downloader/headers/req")
	headerDropMeter    = metrics.NewMeter("shf/downloader/headers/drop")
	headerTimeoutMeter = metrics.NewMeter("shf/downloader/headers/timeout")

	bodyInMeter      = metrics.NewMeter("shf/downloader/bodies/in")
	bodyReqTimer     = metrics.NewTimer("shf/downloader/bodies/req")
	bodyDropMeter    = metrics.NewMeter("shf/downloader/bodies/drop")
	bodyTimeoutMeter = metrics.NewMeter("shf/downloader/bodies/timeout")

	receiptInMeter      = metrics.NewMeter("shf/downloader/receipts/in")
	receiptReqTimer     = metrics.NewTimer("shf/downloader/receipts/req")
	receiptDropMeter    = metrics.NewMeter("shf/downloader/receipts/drop")
	receiptTimeoutMeter = metrics.NewMeter("shf/downloader/receipts/timeout")

	stateInMeter      = metrics.NewMeter("shf/downloader/states/in")
	stateReqTimer     = metrics.NewTimer("shf/downloader/states/req")
	stateDropMeter    = metrics.NewMeter("shf/downloader/states/drop")
	stateTimeoutMeter = metrics.NewMeter("shf/downloader/states/timeout")
)
