package adapter

import (
	"fmt"

	bcore "github.com/ethereum/go-ethereum/consensus/bihs/core"
	ocommon "github.com/ethereum/go-ethereum/consensus/bihs/serialization"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type Block types.Block

func DefaultBlock() bcore.Block {
	return &Block{}
}

func (b *Block) TimeMil() uint64 {
	return (*types.Block)(b).Header().Time * 1000
}

func (b *Block) Height() uint64 {
	return (*types.Block)(b).Header().Number.Uint64()
}

func (b *Block) Hash() bcore.Hash {
	hash := (*types.Block)(b).Header().Hash()
	return hash[:]
}

func (b *Block) Empty() bool {
	return len((*types.Block)(b).Transactions()) == 0
}

func (b *Block) Serialize(sink *ocommon.ZeroCopySink) {
	// (*types.Block)(b).EncodeRLP()
	bytes, err := rlp.EncodeToBytes((*types.Block)(b))
	if err != nil {
		panic(fmt.Sprintf("rlp.EncodeToBytes failed:%v", err))
	}
	sink.WriteVarBytes(bytes)
}

func (b *Block) Deserialize(source *ocommon.ZeroCopySource) error {
	bytes, err := source.ReadVarBytes()
	if err != nil {
		return fmt.Errorf("Block.Deserialize source.ReadVarBytes failed:%v", err)
	}
	return rlp.DecodeBytes(bytes, (*types.Block)(b))
}

func (b *Block) Header() *types.Header {
	return (*types.Block)(b).Header()
}

func (b *Block) WithSeal(header *types.Header) *types.Block {
	return (*types.Block)(b).WithSeal(header)
}
