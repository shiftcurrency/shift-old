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

// Package shf implements the Shift protocol.
package shf

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/shiftcurrency/ethash"
	"github.com/shiftcurrency/shift/accounts"
	"github.com/shiftcurrency/shift/common"
	"github.com/shiftcurrency/shift/common/compiler"
	"github.com/shiftcurrency/shift/common/httpclient"
	"github.com/shiftcurrency/shift/core"
	"github.com/shiftcurrency/shift/core/state"
	"github.com/shiftcurrency/shift/core/types"
	"github.com/shiftcurrency/shift/core/vm"
	"github.com/shiftcurrency/shift/crypto"
	"github.com/shiftcurrency/shift/shf/downloader"
	"github.com/shiftcurrency/shift/ethdb"
	"github.com/shiftcurrency/shift/event"
	"github.com/shiftcurrency/shift/logger"
	"github.com/shiftcurrency/shift/logger/glog"
	"github.com/shiftcurrency/shift/miner"
	"github.com/shiftcurrency/shift/p2p"
	"github.com/shiftcurrency/shift/p2p/discover"
	"github.com/shiftcurrency/shift/p2p/nat"
	"github.com/shiftcurrency/shift/rlp"
	"github.com/shiftcurrency/shift/whisper"
	"github.com/shiftcurrency/shift/sqldb"
    "github.com/shiftcurrency/shift/pow"
)

const (
	epochLength    = 30000
	ethashRevision = 23

	autoDAGcheckInterval = 10 * time.Hour
	autoDAGepochHeight   = epochLength / 2
)

var (
	jsonlogger = logger.NewJsonLogger()

	datadirInUseErrnos = map[uint]bool{11: true, 32: true, 35: true}
	portInUseErrRE     = regexp.MustCompile("address already in use")

	defaultBootNodes = []*discover.Node{
        // London
        discover.MustParseNode("enode://4c8635f108dae8a997697d9c22ddca36969e7f9bc57d9fc01102d7e7d9633231331ae7f7307aceb1aa19130b5bdd4afe397db616c76e7ffc1c69302ba0d09a39@45.32.182.61:53900"),
        // New York
        discover.MustParseNode("enode://fe2a7c2015a8370c6a0d11e7fa6d00000d6428f9576b875ee65c98ff0a0311ed6386cb3eb524806837d843158a7477bd782e0e4870e2ff63b697a088e14891ce@104.207.129.159:53900"),
        // Paris
        discover.MustParseNode("enode://80d0ce5c992f8cc83cdbfd6d832b2dff2e82fee1f8b58762cd858eaacfcc99d5a8a837648bd28a2d508cc1da305c15cf4e531546034ed1a8ccd07ff51a71abd6@108.61.177.0:53900"),
        // Seattle
        discover.MustParseNode("enode://f019da062a635a4e9e89ec93edc7ca11c06fdfec0574f1cb001126a82dc6ffa6ca05f924a683934ff5d01fc5d4b0ac9507349a945c97121b2a355d39b1781cd7@104.238.157.156:53900"),
	}

	defaultTestNetBootNodes = []*discover.Node{
	}

	staticNodes  = "static-nodes.json"  // Path within <datadir> to search for the static node list
	trustedNodes = "trusted-nodes.json" // Path within <datadir> to search for the trusted node list
)

type Config struct {
	DevMode bool
	TestNet bool

	Name         string
	NetworkId    int
	GenesisFile  string
	GenesisBlock *types.Block // used by block tests
	FastSync     bool
	Olympic      bool

	BlockChainVersion  int
	SkipBcVersionCheck bool // e.g. blockchain export
	DatabaseCache      int

	DataDir   string
	LogFile   string
	Verbosity int
	VmDebug   bool
	NatSpec   bool
	DocRoot   string
	AutoDAG   bool
	PowTest   bool
	ExtraData []byte

	MaxPeers        int
	MaxPendingPeers int
	Discovery       bool
	Port            string

	// Space-separated list of discovery node URLs
	BootNodes string

	// This key is used to identify the node on the network.
	// If nil, an ephemeral key is used.
	NodeKey *ecdsa.PrivateKey

	NAT  nat.Interface
	Shh  bool
	Dial bool

	Shiftbase      common.Address
	GasPrice       *big.Int
	MinerThreads   int
	AccountManager *accounts.Manager
	SolcPath       string

	GpoMinGasPrice          *big.Int
	GpoMaxGasPrice          *big.Int
	GpoFullBlockRatio       int
	GpobaseStepDown         int
	GpobaseStepUp           int
	GpobaseCorrectionFactor int

	// NewDB is used to create databases.
	// If nil, the default is to create leveldb databases on disk.
	NewDB func(path string) (ethdb.Database, error)
}

func (cfg *Config) parseBootNodes() []*discover.Node {
	if cfg.BootNodes == "" {
		if cfg.TestNet {
			return defaultTestNetBootNodes
		}

		return defaultBootNodes
	}
	var ns []*discover.Node
	for _, url := range strings.Split(cfg.BootNodes, " ") {
		if url == "" {
			continue
		}
		n, err := discover.ParseNode(url)
		if err != nil {
			glog.V(logger.Error).Infof("Bootstrap URL %s: %v\n", url, err)
			continue
		}
		ns = append(ns, n)
	}
	return ns
}

// parseNodes parses a list of discovery node URLs loaded from a .json file.
func (cfg *Config) parseNodes(file string) []*discover.Node {
	// Short circuit if no node config is present
	path := filepath.Join(cfg.DataDir, file)
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	// Load the nodes from the config file
	blob, err := ioutil.ReadFile(path)
	if err != nil {
		glog.V(logger.Error).Infof("Failed to access nodes: %v", err)
		return nil
	}
	nodelist := []string{}
	if err := json.Unmarshal(blob, &nodelist); err != nil {
		glog.V(logger.Error).Infof("Failed to load nodes: %v", err)
		return nil
	}
	// Interpret the list as a discovery node array
	var nodes []*discover.Node
	for _, url := range nodelist {
		if url == "" {
			continue
		}
		node, err := discover.ParseNode(url)
		if err != nil {
			glog.V(logger.Error).Infof("Node URL %s: %v\n", url, err)
			continue
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func (cfg *Config) nodeKey() (*ecdsa.PrivateKey, error) {
	// use explicit key from command line args if set
	if cfg.NodeKey != nil {
		return cfg.NodeKey, nil
	}
	// use persistent key if present
	keyfile := filepath.Join(cfg.DataDir, "nodekey")
	key, err := crypto.LoadECDSA(keyfile)
	if err == nil {
		return key, nil
	}
	// no persistent key, generate and store a new one
	if key, err = crypto.GenerateKey(); err != nil {
		return nil, fmt.Errorf("could not generate server key: %v", err)
	}
	if err := crypto.SaveECDSA(keyfile, key); err != nil {
		glog.V(logger.Error).Infoln("could not persist nodekey: ", err)
	}
	return key, nil
}

type Shift struct {
	// Channel for shutting down the shift
	shutdownChan chan bool

	// DB interfaces
	chainDb ethdb.Database // Block chain database
	dappDb  ethdb.Database // Dapp database
    sqlDB   *sqldb.SQLDB

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
    chainManager    *pow.ChainManager
	accountManager  *accounts.Manager
	whisper         *whisper.Whisper
	pow             *ethash.Ethash
	protocolManager *ProtocolManager
	SolcPath        string
	solc            *compiler.Solidity

	GpoMinGasPrice          *big.Int
	GpoMaxGasPrice          *big.Int
	GpoFullBlockRatio       int
	GpobaseStepDown         int
	GpobaseStepUp           int
	GpobaseCorrectionFactor int

	httpclient *httpclient.HTTPClient

	net      *p2p.Server
	eventMux *event.TypeMux
	miner    *miner.Miner

	// logger logger.LogSystem

	Mining        bool
	MinerThreads  int
	NatSpec       bool
	DataDir       string
	AutoDAG       bool
	PowTest       bool
	autodagquit   chan bool
	shiftbase     common.Address
	clientVersion string
	netVersionId  int
	shhVersionId  int
}

func New(config *Config) (*Shift, error) {
	logger.New(config.DataDir, config.LogFile, config.Verbosity)

	// Let the database take 3/4 of the max open files (TODO figure out a way to get the actual limit of the open files)
	const dbCount = 3
	ethdb.OpenFileLimit = 128 / (dbCount + 1)

	newdb := config.NewDB
	if newdb == nil {
		newdb = func(path string) (ethdb.Database, error) { return ethdb.NewLDBDatabase(path, config.DatabaseCache) }
	}

	// Open the chain database and perform any upgrades needed
	chainDb, err := newdb(filepath.Join(config.DataDir, "chaindata"))
	if err != nil {
		if errno, ok := err.(syscall.Errno); ok && datadirInUseErrnos[uint(errno)] {
			err = fmt.Errorf("%v (check if another instance of geth is already running with the same data directory '%s')", err, config.DataDir)
		}
		return nil, fmt.Errorf("blockchain db err: %v", err)
	}
	if db, ok := chainDb.(*ethdb.LDBDatabase); ok {
		db.Meter("eth/db/chaindata/")
	}
	if err := upgradeChainDatabase(chainDb); err != nil {
		return nil, err
	}
	if err := addMipmapBloomBins(chainDb); err != nil {
		return nil, err
	}

	dappDb, err := newdb(filepath.Join(config.DataDir, "dapp"))
	if err != nil {
		if errno, ok := err.(syscall.Errno); ok && datadirInUseErrnos[uint(errno)] {
			err = fmt.Errorf("%v (check if another instance of geth is already running with the same data directory '%s')", err, config.DataDir)
		}
		return nil, fmt.Errorf("dapp db err: %v", err)
	}
	if db, ok := dappDb.(*ethdb.LDBDatabase); ok {
		db.Meter("eth/db/dapp/")
	}

	nodeDb := filepath.Join(config.DataDir, "nodes")
	glog.V(logger.Info).Infof("Protocol Versions: %v, Network Id: %v", ProtocolVersions, config.NetworkId)

	if len(config.GenesisFile) > 0 {
		fr, err := os.Open(config.GenesisFile)
		if err != nil {
			return nil, err
		}

		block, err := core.WriteGenesisBlock(chainDb, fr)
		if err != nil {
			return nil, err
		}
		glog.V(logger.Info).Infof("Successfully wrote genesis block. New genesis hash = %x\n", block.Hash())
	}

	// different modes
	switch {
	case config.Olympic:
		glog.V(logger.Error).Infoln("Starting Olympic network")
		fallthrough
	case config.DevMode:
		_, err := core.WriteOlympicGenesisBlock(chainDb, 42)
		if err != nil {
			return nil, err
		}
	case config.TestNet:
		state.StartingNonce = 1048576 // (2**20)
		_, err := core.WriteTestNetGenesisBlock(chainDb, 0x6d6f7264656e)
		if err != nil {
			return nil, err
		}
	}
	// This is for testing only.
	if config.GenesisBlock != nil {
		core.WriteTd(chainDb, config.GenesisBlock.Hash(), config.GenesisBlock.Difficulty())
		core.WriteBlock(chainDb, config.GenesisBlock)
		core.WriteCanonicalHash(chainDb, config.GenesisBlock.Hash(), config.GenesisBlock.NumberU64())
		core.WriteHeadBlockHash(chainDb, config.GenesisBlock.Hash())
	}

	if !config.SkipBcVersionCheck {
		b, _ := chainDb.Get([]byte("BlockchainVersion"))
		bcVersion := int(common.NewValue(b).Uint())
		if bcVersion != config.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run gshift upgradedb.\n", bcVersion, config.BlockChainVersion)
		}
		saveBlockchainVersion(chainDb, config.BlockChainVersion)
	}
	glog.V(logger.Info).Infof("Blockchain DB Version: %d", config.BlockChainVersion)

	shf := &Shift{
		shutdownChan:            make(chan bool),
		chainDb:                 chainDb,
		dappDb:                  dappDb,
		eventMux:                &event.TypeMux{},
		accountManager:          config.AccountManager,
		DataDir:                 config.DataDir,
		shiftbase:               config.Shiftbase,
		clientVersion:           config.Name, // TODO should separate from Name
		netVersionId:            config.NetworkId,
		NatSpec:                 config.NatSpec,
		MinerThreads:            config.MinerThreads,
		SolcPath:                config.SolcPath,
		AutoDAG:                 config.AutoDAG,
		PowTest:                 config.PowTest,
		GpoMinGasPrice:          config.GpoMinGasPrice,
		GpoMaxGasPrice:          config.GpoMaxGasPrice,
		GpoFullBlockRatio:       config.GpoFullBlockRatio,
		GpobaseStepDown:         config.GpobaseStepDown,
		GpobaseStepUp:           config.GpobaseStepUp,
		GpobaseCorrectionFactor: config.GpobaseCorrectionFactor,
		httpclient:              httpclient.New(config.DocRoot),
	}

	shf.sqlDB, err = sqldb.NewSQLiteDatabase(filepath.Join(config.DataDir, "sql.db"))
	if err != nil {
		return nil, err
	}


	if config.PowTest {
		glog.V(logger.Info).Infof("ethash used in test mode")
		shf.pow, err = ethash.NewForTesting()
		if err != nil {
			return nil, err
		}
	} else {
		shf.pow = ethash.New()
	}
	//genesis := core.GenesisBlock(uint64(config.GenesisNonce), stateDb)
	shf.blockchain, err = core.NewBlockChain(chainDb, shf.sqlDB, shf.pow, shf.EventMux())
	if err != nil {
		if err == core.ErrNoGenesis {
			return nil, fmt.Errorf(`Genesis block not found. Please supply a genesis block with the "--genesis /path/to/file" argument`)
		}
		return nil, err
	}
	newPool := core.NewTxPool(shf.EventMux(), shf.blockchain.State, shf.blockchain.GasLimit)
	shf.txPool = newPool

	if shf.protocolManager, err = NewProtocolManager(config.FastSync, config.NetworkId, shf.eventMux, shf.txPool, shf.pow, shf.blockchain, chainDb); err != nil {
		return nil, err
	}
	shf.miner = miner.New(shf, shf.EventMux(), shf.pow)
	shf.miner.SetGasPrice(config.GasPrice)
	shf.miner.SetExtra(config.ExtraData)

	if config.Shh {
		shf.whisper = whisper.New()
		shf.shhVersionId = int(shf.whisper.Version())
	}

	netprv, err := config.nodeKey()
	if err != nil {
		return nil, err
	}
	protocols := append([]p2p.Protocol{}, shf.protocolManager.SubProtocols...)
	if config.Shh {
		protocols = append(protocols, shf.whisper.Protocol())
	}
	shf.net = &p2p.Server{
		PrivateKey:      netprv,
		Name:            config.Name,
		MaxPeers:        config.MaxPeers,
		MaxPendingPeers: config.MaxPendingPeers,
		Discovery:       config.Discovery,
		Protocols:       protocols,
		NAT:             config.NAT,
		NoDial:          !config.Dial,
		BootstrapNodes:  config.parseBootNodes(),
		StaticNodes:     config.parseNodes(staticNodes),
		TrustedNodes:    config.parseNodes(trustedNodes),
		NodeDatabase:    nodeDb,
	}
	if len(config.Port) > 0 {
		shf.net.ListenAddr = ":" + config.Port
	}

	vm.Debug = config.VmDebug
	shf.sqlDB.Refresh(shf.blockchain)

	return shf, nil
}

// Network retrieves the underlying P2P network server. This should eventually
// be moved out into a protocol independent package, but for now use an accessor.
func (s *Shift) Network() *p2p.Server {
	return s.net
}

func (s *Shift) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Shift) Shiftbase() (eb common.Address, err error) {
	eb = s.shiftbase
	if (eb == common.Address{}) {
		addr, e := s.AccountManager().AddressByIndex(0)
		if e != nil {
			err = fmt.Errorf("shiftbase address must be explicitly specified")
		}
		eb = common.HexToAddress(addr)
	}
	return
}

// set in js console via admin interface or wrapper from cli flags
func (self *Shift) SetShiftbase(shiftbase common.Address) {
	self.shiftbase = shiftbase
	self.miner.SetShiftbase(shiftbase)
}


func (s *Shift) StopMining()         { s.miner.Stop() }
func (s *Shift) IsMining() bool      { return s.miner.Mining() }
func (s *Shift) Miner() *miner.Miner { return s.miner }

// func (s *Shift) Logger() logger.LogSystem             { return s.logger }
func (s *Shift) Name() string                       { return s.net.Name }
func (s *Shift) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *Shift) ChainManager()     *pow.ChainManager { return s.chainManager }
func (s *Shift) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *Shift) SQLDB() *sqldb.SQLDB     	        { return s.sqlDB }
func (s *Shift) TxPool() *core.TxPool               { return s.txPool }
func (s *Shift) Whisper() *whisper.Whisper          { return s.whisper }
func (s *Shift) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Shift) ChainDb() ethdb.Database            { return s.chainDb }
func (s *Shift) DappDb() ethdb.Database             { return s.dappDb }
func (s *Shift) IsListening() bool                  { return true } // Always listening
func (s *Shift) PeerCount() int                     { return s.net.PeerCount() }
func (s *Shift) Peers() []*p2p.Peer                 { return s.net.Peers() }
func (s *Shift) MaxPeers() int                      { return s.net.MaxPeers }
func (s *Shift) ClientVersion() string              { return s.clientVersion }
func (s *Shift) EthVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Shift) NetVersion() int                    { return s.netVersionId }
func (s *Shift) ShhVersion() int                    { return s.shhVersionId }
func (s *Shift) Downloader() *downloader.Downloader { return s.protocolManager.downloader }

// Start the ethereum
func (s *Shift) Start() error {
	jsonlogger.LogJson(&logger.LogStarting{
		ClientString:    s.net.Name,
		ProtocolVersion: s.EthVersion(),
	})
	err := s.net.Start()
	if err != nil {
		if portInUseErrRE.MatchString(err.Error()) {
			err = fmt.Errorf("%v (possibly another instance of geth is using the same port)", err)
		}
		return err
	}

	if s.AutoDAG {
		s.StartAutoDAG()
	}

	s.protocolManager.Start()

	if s.whisper != nil {
		s.whisper.Start()
	}

	glog.V(logger.Info).Infoln("Server started")
	return nil
}

func (s *Shift) StartForTest() {
	jsonlogger.LogJson(&logger.LogStarting{
		ClientString:    s.net.Name,
		ProtocolVersion: s.EthVersion(),
	})
}

// AddPeer connects to the given node and maintains the connection until the
// server is shut down. If the connection fails for any reason, the server will
// attempt to reconnect the peer.
func (self *Shift) AddPeer(nodeURL string) error {
	n, err := discover.ParseNode(nodeURL)
	if err != nil {
		return fmt.Errorf("invalid node URL: %v", err)
	}
	self.net.AddPeer(n)
	return nil
}

func (s *Shift) Stop() {
	s.net.Stop()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()
	s.eventMux.Stop()
	if s.whisper != nil {
		s.whisper.Stop()
	}
	s.StopAutoDAG()

	s.chainDb.Close()
	s.dappDb.Close()
	close(s.shutdownChan)
}

// This function will wait for a shutdown and resumes main thread execution
func (s *Shift) WaitForShutdown() {
	<-s.shutdownChan
}

// StartAutoDAG() spawns a go routine that checks the DAG every autoDAGcheckInterval
// by default that is 10 times per epoch
// in epoch n, if we past autoDAGepochHeight within-epoch blocks,
// it calls ethash.MakeDAG  to pregenerate the DAG for the next epoch n+1
// if it does not exist yet as well as remove the DAG for epoch n-1
// the loop quits if autodagquit channel is closed, it can safely restart and
// stop any number of times.
// For any more sophisticated pattern of DAG generation, use CLI subcommand
// makedag
func (self *Shift) StartAutoDAG() {
	if self.autodagquit != nil {
		return // already started
	}
	go func() {
		glog.V(logger.Info).Infof("Automatic pregeneration of ethash DAG ON (ethash dir: %s)", ethash.DefaultDir)
		var nextEpoch uint64
		timer := time.After(0)
		self.autodagquit = make(chan bool)
		for {
			select {
			case <-timer:
				glog.V(logger.Info).Infof("checking DAG (ethash dir: %s)", ethash.DefaultDir)
				currentBlock := self.BlockChain().CurrentBlock().NumberU64()
				thisEpoch := currentBlock / epochLength
				if nextEpoch <= thisEpoch {
					if currentBlock%epochLength > autoDAGepochHeight {
						if thisEpoch > 0 {
							previousDag, previousDagFull := dagFiles(thisEpoch - 1)
							os.Remove(filepath.Join(ethash.DefaultDir, previousDag))
							os.Remove(filepath.Join(ethash.DefaultDir, previousDagFull))
							glog.V(logger.Info).Infof("removed DAG for epoch %d (%s)", thisEpoch-1, previousDag)
						}
						nextEpoch = thisEpoch + 1
						dag, _ := dagFiles(nextEpoch)
						if _, err := os.Stat(dag); os.IsNotExist(err) {
							glog.V(logger.Info).Infof("Pregenerating DAG for epoch %d (%s)", nextEpoch, dag)
							err := ethash.MakeDAG(nextEpoch*epochLength, "") // "" -> ethash.DefaultDir
							if err != nil {
								glog.V(logger.Error).Infof("Error generating DAG for epoch %d (%s)", nextEpoch, dag)
								return
							}
						} else {
							glog.V(logger.Error).Infof("DAG for epoch %d (%s)", nextEpoch, dag)
						}
					}
				}
				timer = time.After(autoDAGcheckInterval)
			case <-self.autodagquit:
				return
			}
		}
	}()
}

// stopAutoDAG stops automatic DAG pregeneration by quitting the loop
func (self *Shift) StopAutoDAG() {
	if self.autodagquit != nil {
		close(self.autodagquit)
		self.autodagquit = nil
	}
	glog.V(logger.Info).Infof("Automatic pregeneration of ethash DAG OFF (ethash dir: %s)", ethash.DefaultDir)
}


// HTTPClient returns the light http client used for fetching offchain docs
// (natspec, source for verification)
func (self *Shift) HTTPClient() *httpclient.HTTPClient {
	return self.httpclient
}

func (self *Shift) Solc() (*compiler.Solidity, error) {
	var err error
	if self.solc == nil {
		self.solc, err = compiler.New(self.SolcPath)
	}
	return self.solc, err
}

// set in js console via admin interface or wrapper from cli flags
func (self *Shift) SetSolc(solcPath string) (*compiler.Solidity, error) {
	self.SolcPath = solcPath
	self.solc = nil
	return self.Solc()
}

// dagFiles(epoch) returns the two alternative DAG filenames (not a path)
// 1) <revision>-<hex(seedhash[8])> 2) full-R<revision>-<hex(seedhash[8])>
func dagFiles(epoch uint64) (string, string) {
	seedHash, _ := ethash.GetSeedHash(epoch * epochLength)
	dag := fmt.Sprintf("full-R%d-%x", ethashRevision, seedHash[:8])
	return dag, "full-R" + dag
}

func saveBlockchainVersion(db ethdb.Database, bcVersion int) {
	d, _ := db.Get([]byte("BlockchainVersion"))
	blockchainVersion := common.NewValue(d).Uint()

	if blockchainVersion == 0 {
		db.Put([]byte("BlockchainVersion"), common.NewValue(bcVersion).Bytes())
	}
}

// upgradeChainDatabase ensures that the chain database stores block split into
// separate header and body entries.
func upgradeChainDatabase(db ethdb.Database) error {
	// Short circuit if the head block is stored already as separate header and body
	data, err := db.Get([]byte("LastBlock"))
	if err != nil {
		return nil
	}
	head := common.BytesToHash(data)

	if block := core.GetBlockByHashOld(db, head); block == nil {
		return nil
	}
	// At least some of the database is still the old format, upgrade (skip the head block!)
	glog.V(logger.Info).Info("Old database detected, upgrading...")

	if db, ok := db.(*ethdb.LDBDatabase); ok {
		blockPrefix := []byte("block-hash-")
		for it := db.NewIterator(); it.Next(); {
			// Skip anything other than a combined block
			if !bytes.HasPrefix(it.Key(), blockPrefix) {
				continue
			}
			// Skip the head block (merge last to signal upgrade completion)
			if bytes.HasSuffix(it.Key(), head.Bytes()) {
				continue
			}
			// Load the block, split and serialize (order!)
			block := core.GetBlockByHashOld(db, common.BytesToHash(bytes.TrimPrefix(it.Key(), blockPrefix)))

			if err := core.WriteTd(db, block.Hash(), block.DeprecatedTd()); err != nil {
				return err
			}
			if err := core.WriteBody(db, block.Hash(), &types.Body{block.Transactions(), block.Uncles()}); err != nil {
				return err
			}
			if err := core.WriteHeader(db, block.Header()); err != nil {
				return err
			}
			if err := db.Delete(it.Key()); err != nil {
				return err
			}
		}
		// Lastly, upgrade the head block, disabling the upgrade mechanism
		current := core.GetBlockByHashOld(db, head)

		if err := core.WriteTd(db, current.Hash(), current.DeprecatedTd()); err != nil {
			return err
		}
		if err := core.WriteBody(db, current.Hash(), &types.Body{current.Transactions(), current.Uncles()}); err != nil {
			return err
		}
		if err := core.WriteHeader(db, current.Header()); err != nil {
			return err
		}
	}
	return nil
}

func addMipmapBloomBins(db ethdb.Database) (err error) {
	const mipmapVersion uint = 2

	// check if the version is set. We ignore data for now since there's
	// only one version so we can easily ignore it for now
	var data []byte
	data, _ = db.Get([]byte("setting-mipmap-version"))
	if len(data) > 0 {
		var version uint
		if err := rlp.DecodeBytes(data, &version); err == nil && version == mipmapVersion {
			return nil
		}
	}

	defer func() {
		if err == nil {
			var val []byte
			val, err = rlp.EncodeToBytes(mipmapVersion)
			if err == nil {
				err = db.Put([]byte("setting-mipmap-version"), val)
			}
			return
		}
	}()
	latestBlock := core.GetBlock(db, core.GetHeadBlockHash(db))
	if latestBlock == nil { // clean database
		return
	}

	tstart := time.Now()
	glog.V(logger.Info).Infoln("upgrading db log bloom bins")
	for i := uint64(0); i <= latestBlock.NumberU64(); i++ {
		hash := core.GetCanonicalHash(db, i)
		if (hash == common.Hash{}) {
			return fmt.Errorf("chain db corrupted. Could not find block %d.", i)
		}
		core.WriteMipmapBloom(db, i, core.GetBlockReceipts(db, hash))
	}
	glog.V(logger.Info).Infoln("upgrade completed in", time.Since(tstart))
	return nil
}
