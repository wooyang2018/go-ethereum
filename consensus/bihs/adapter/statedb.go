package adapter

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/consensus"
	bcore "github.com/ethereum/go-ethereum/consensus/bihs/core"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	ocommon "github.com/ontio/ontology/common"
)

type StateDB struct {
	sync.RWMutex
	chain                  *core.BlockChain
	gov                    *Governance
	prepareEmptyHeaderFunc func() *types.Header
	saveBlockFunc          func(block *types.Block)
	verifyHeaderFunc       func(chain consensus.ChainHeaderReader, header *types.Header, seal bool) error
	heightSubs             []bcore.HeightChangeSub
}

func NewStateDB(chain *core.BlockChain, gov *Governance, prepareEmptyHeaderFunc func() *types.Header, saveBlockFunc func(block *types.Block), verifyHeaderFunc func(chain consensus.ChainHeaderReader, header *types.Header, seal bool) error) *StateDB {
	db := &StateDB{
		chain:                  chain,
		gov:                    gov,
		prepareEmptyHeaderFunc: prepareEmptyHeaderFunc,
		verifyHeaderFunc:       verifyHeaderFunc,
		saveBlockFunc:          saveBlockFunc,
	}

	return db
}

func (db *StateDB) StoreBlock(blk bcore.Block, commitQC *bcore.QC) error {
	if commitQC.Type != bcore.MTPreCommit {
		return fmt.Errorf("invalid type for commitQC")
	}

	sink := ocommon.NewZeroCopySink(nil)
	commitQC.SerializeForHeader(sink)

	block := blk.(*Block)
	header := block.Header()
	header.Extra = sink.Bytes()

	db.saveBlockFunc(block.WithSeal(header))
	return nil
}

func (db *StateDB) Validate(blk bcore.Block) (err error) {
	block, ok := blk.(*Block)
	if !ok {
		err = fmt.Errorf("invalid block")
		return
	}
	err = db.verifyHeaderFunc(db.chain, block.Header(), false)
	if err != nil {
		return
	}
	return preExecuteBlock(db.chain, (*types.Block)(block))
}

func preExecuteBlock(bc *core.BlockChain, block *types.Block) error {
	parent := bc.GetBlockByHash(block.ParentHash())
	statedb, err := bc.StateAt(parent.Root())
	if err != nil {
		return err
	}

	txHash := types.DeriveSha(block.Transactions(), trie.NewStackTrie(nil))
	if block.TxHash() != txHash {
		return fmt.Errorf("invalid txHash")
	}
	receipts, _, usedGas, err := bc.Processor().Process(block, statedb, *bc.GetVMConfig())
	if err != nil {
		return err
	}
	if err := bc.Validator().ValidateState(block, statedb, receipts, usedGas); err != nil {
		return err
	}
	return nil
}

func (db *StateDB) MakeBlock(height uint64, mustEmpty bool) (bcore.Block, error) {
	emptyHeader := db.prepareEmptyHeaderFunc()
	if emptyHeader == nil {
		return nil, fmt.Errorf("prepareEmptyHeaderFunc failed")
	}
	if emptyHeader.Number.Uint64() != height {
		return nil, fmt.Errorf("emptyHeader wrong height, expect:%d got:%d", height, emptyHeader.Number.Uint64())
	}
	return (*Block)(types.NewBlock(emptyHeader, nil, nil, nil, trie.NewStackTrie(nil))), nil
}

func (db *StateDB) Height() uint64 {
	return db.chain.CurrentHeader().Number.Uint64()
}

func (db *StateDB) SubscribeHeightChange(sub bcore.HeightChangeSub) {
	db.Lock()
	defer db.Unlock()

	db.heightSubs = append(db.heightSubs, sub)
}

func (db *StateDB) HeightChanged() {
	db.RLock()
	heightSubs := db.heightSubs
	db.RUnlock()

	for _, sub := range heightSubs {
		sub.HeightChanged()
	}
}

func (db *StateDB) UnSubscribeHeightChange(sub bcore.HeightChangeSub) {
	db.Lock()
	defer db.Unlock()

	count := len(db.heightSubs)
	for i, subed := range db.heightSubs {
		if subed == sub {
			db.heightSubs[count-1], db.heightSubs[i] = db.heightSubs[i], db.heightSubs[count-1]
			db.heightSubs = db.heightSubs[0 : count-1]
			return
		}
	}
}

func (db *StateDB) ValidatorIndex(height uint64, peer bcore.ID) int {
	return db.gov.ValidatorIndex(height, peer)
}

func (db *StateDB) SelectLeader(height, view uint64) bcore.ID {
	return db.gov.SelectLeader(height, view)
}

func (db *StateDB) ValidatorCount(height uint64) int32 {
	return db.gov.ValidatorCount(height)
}

func (db *StateDB) ValidatorIDs(height uint64) []bcore.ID {
	return db.gov.ValidatorIDs(height)
}

func (db *StateDB) PKs(height uint64, bitmap []byte) interface{} {
	panic("not used")
}
