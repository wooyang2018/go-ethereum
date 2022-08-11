package adapter

import (
	"github.com/ethereum/go-ethereum/common"
	bcore "github.com/ethereum/go-ethereum/consensus/bihs/core"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/log"
)

// this package is for test purpose only

type Governance struct {
	chain *core.BlockChain
}

func NewGov(chain *core.BlockChain) *Governance {
	return &Governance{chain: chain}
}

var validators = []common.Address{
	common.HexToAddress("0xde5B5Dd07C7EE63712b334EcD59E3FA173E6d56E"),
	common.HexToAddress("0xD642f9b4c28F6bA62126144B7E26e8Cf85CB2d3a"),
}

func (g *Governance) ValidatorP2PAddrs(height uint64) []common.Address {
	return validators
}

func (g *Governance) ValidatorP2PAddr(account common.Address) common.Address {
	return account
}

func (g *Governance) ValidatorIndex(height uint64, peer bcore.ID) int {
	peerAddr := common.BytesToAddress(peer)
	log.Info("ValidatorIndex", "peerAddr", peerAddr)
	for i, addr := range validators {
		if peerAddr == addr {
			return i
		}
	}

	log.Info("ValidatorIndex", "peerAddr", peerAddr, "idx", -1)
	return -1
}

func (g *Governance) SelectLeader(height, view uint64) bcore.ID {
	return validators[(height+view)%uint64(len(validators))][:]
}

func (g *Governance) ValidatorCount(height uint64) int32 {
	return int32(len(validators))
}

func (g *Governance) ValidatorIDs(height uint64) (result []bcore.ID) {
	for _, val := range validators {
		val := val
		result = append(result, val[:])
	}
	return
}
