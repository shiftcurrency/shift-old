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

package shf

import (
	"math/big"
	"math/rand"
	"sync"
	"github.com/shiftcurrency/shift/core"
	"github.com/shiftcurrency/shift/core/types"
	"github.com/shiftcurrency/shift/logger"
	"github.com/shiftcurrency/shift/logger/glog"
)

const (
	gpoProcessPastBlocks = 100

	// for testing
	gpoDefaultBaseCorrectionFactor = 110
	gpoDefaultMinNrgPrice          = 10000000000000
)

type blockPriceInfo struct {
	baseNrgPrice *big.Int
}

// NrgPriceOracle recommends nrg prices based on the content of recent
// blocks.
type NrgPriceOracle struct {
	shf           *Shift
	initOnce      sync.Once
	minPrice      *big.Int
	lastBaseMutex sync.Mutex
	lastBase      *big.Int

	// state of listenLoop
	blocks                        map[uint64]*blockPriceInfo
	firstProcessed, lastProcessed uint64
	minBase                       *big.Int
}

// NewNrgPriceOracle returns a new oracle.
func NewNrgPriceOracle(shf *Shift) *NrgPriceOracle {
	minprice := shf.GpoMinNrgPrice
	if minprice == nil {
		minprice = big.NewInt(gpoDefaultMinNrgPrice)
	}
	minbase := new(big.Int).Mul(minprice, big.NewInt(100))
	if shf.GpobaseCorrectionFactor > 0 {
		minbase = minbase.Div(minbase, big.NewInt(int64(shf.GpobaseCorrectionFactor)))
	}
	return &NrgPriceOracle{
		shf:      shf,
		blocks:   make(map[uint64]*blockPriceInfo),
		minBase:  minbase,
		minPrice: minprice,
		lastBase: minprice,
	}
}

func (gpo *NrgPriceOracle) init() {
	gpo.initOnce.Do(func() {
		gpo.processPastBlocks(gpo.shf.BlockChain())
		go gpo.listenLoop()
	})
}

func (self *NrgPriceOracle) processPastBlocks(chain *core.BlockChain) {
	last := int64(-1)
	cblock := chain.CurrentBlock()
	if cblock != nil {
		last = int64(cblock.NumberU64())
	}
	first := int64(0)
	if last > gpoProcessPastBlocks {
		first = last - gpoProcessPastBlocks
	}
	self.firstProcessed = uint64(first)
	for i := first; i <= last; i++ {
		block := chain.GetBlockByNumber(uint64(i))
		if block != nil {
			self.processBlock(block)
		}
	}

}

func (self *NrgPriceOracle) listenLoop() {
	events := self.shf.EventMux().Subscribe(core.ChainEvent{}, core.ChainSplitEvent{})
	defer events.Unsubscribe()

	for event := range events.Chan() {
		switch event := event.Data.(type) {
		case core.ChainEvent:
			self.processBlock(event.Block)
		case core.ChainSplitEvent:
			self.processBlock(event.Block)
		}
	}
}

func (self *NrgPriceOracle) processBlock(block *types.Block) {
	i := block.NumberU64()
	if i > self.lastProcessed {
		self.lastProcessed = i
	}

	lastBase := self.minPrice
	bpl := self.blocks[i-1]
	if bpl != nil {
		lastBase = bpl.baseNrgPrice
	}
	if lastBase == nil {
		return
	}

	var corr int
	lp := self.lowestPrice(block)
	if lp == nil {
		return
	}

	if lastBase.Cmp(lp) < 0 {
		corr = self.shf.GpobaseStepUp
	} else {
		corr = -self.shf.GpobaseStepDown
	}

	crand := int64(corr * (900 + rand.Intn(201)))
	newBase := new(big.Int).Mul(lastBase, big.NewInt(1000000+crand))
	newBase.Div(newBase, big.NewInt(1000000))

	if newBase.Cmp(self.minBase) < 0 {
		newBase = self.minBase
	}

	bpi := self.blocks[i]
	if bpi == nil {
		bpi = &blockPriceInfo{}
		self.blocks[i] = bpi
	}
	bpi.baseNrgPrice = newBase
	self.lastBaseMutex.Lock()
	self.lastBase = newBase
	self.lastBaseMutex.Unlock()

	glog.V(logger.Detail).Infof("Processed block #%v, base price is %v\n", block.NumberU64(), newBase.Int64())
}

// returns the lowers possible price with which a tx was or could have been included
func (self *NrgPriceOracle) lowestPrice(block *types.Block) *big.Int {
	nrgUsed := big.NewInt(0)

	receipts := core.GetBlockReceipts(self.shf.ChainDb(), block.Hash())
	if len(receipts) > 0 {
		if cgu := receipts[len(receipts)-1].CumulativeNrgUsed; cgu != nil {
			nrgUsed = receipts[len(receipts)-1].CumulativeNrgUsed
		}
	}

	if new(big.Int).Mul(nrgUsed, big.NewInt(100)).Cmp(new(big.Int).Mul(block.NrgLimit(),
		big.NewInt(int64(self.shf.GpoFullBlockRatio)))) < 0 {
		// block is not full, could have posted a tx with MinNrgPrice
		return big.NewInt(0)
	}

	txs := block.Transactions()
	if len(txs) == 0 {
		return big.NewInt(0)
	}
	// block is full, find smallest nrgPrice
	minPrice := txs[0].NrgPrice()
	for i := 1; i < len(txs); i++ {
		price := txs[i].NrgPrice()
		if price.Cmp(minPrice) < 0 {
			minPrice = price
		}
	}
	return minPrice
}

// SuggestPrice returns the recommended nrg price.
func (self *NrgPriceOracle) SuggestPrice() *big.Int {
	self.init()
	self.lastBaseMutex.Lock()
	price := new(big.Int).Set(self.lastBase)
	self.lastBaseMutex.Unlock()

	price.Mul(price, big.NewInt(int64(self.shf.GpobaseCorrectionFactor)))
	price.Div(price, big.NewInt(100))
	if price.Cmp(self.minPrice) < 0 {
		price.Set(self.minPrice)
	} else if self.shf.GpoMaxNrgPrice != nil && price.Cmp(self.shf.GpoMaxNrgPrice) > 0 {
		price.Set(self.shf.GpoMaxNrgPrice)
	}
	return price
}
