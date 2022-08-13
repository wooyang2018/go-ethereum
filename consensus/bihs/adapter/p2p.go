package adapter

import (
	"github.com/ethereum/go-ethereum/common"
	bcore "github.com/ethereum/go-ethereum/consensus/bihs/core"
	bser "github.com/ethereum/go-ethereum/consensus/bihs/serialization"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
)

type P2P struct {
	bc    Broadcaster
	chain *core.BlockChain
	gov   *Governance
	ch    chan *bcore.Msg
}

type Broadcaster interface {
	Unicast(target common.Address, msgcode uint64, data interface{})
	Multicast(targets []common.Address, msgcode uint64, data interface{})
}

const chanSize = 20

var ConsensusMsgCode uint64

func NewP2P(bc Broadcaster, chain *core.BlockChain, gov *Governance) *P2P {
	return &P2P{bc: bc, chain: chain, gov: gov, ch: make(chan *bcore.Msg, chanSize)}
}

func (p *P2P) Broadcast(msg *bcore.Msg) {
	sink := bser.NewZeroCopySink(nil)
	msg.Serialize(sink)
	payload := sink.Bytes()
	validators := p.gov.ValidatorP2PAddrs(msg.Height)
	log.Info("Broadcast", "#payload", len(payload), "type", msg.Type, "height", msg.Height, "view", msg.View, "msg hash", msg.Hash())
	p.bc.Multicast(validators, ConsensusMsgCode, payload)
}

func (p *P2P) Send(id bcore.ID, msg *bcore.Msg) {
	target := p.gov.ValidatorP2PAddr(common.BytesToAddress(id))
	sink := bser.NewZeroCopySink(nil)
	msg.Serialize(sink)
	payload := sink.Bytes()
	p.bc.Unicast(target, ConsensusMsgCode, payload)
}

func (p *P2P) MsgCh() <-chan *bcore.Msg {
	return p.ch
}

func (p *P2P) HandleP2pMsg(msg p2p.Msg) (err error) {
	var payload []byte
	if err = msg.Decode(&payload); err != nil {
		return
	}

	var bihsMsg bcore.Msg
	if err = bihsMsg.Deserialize(bser.NewZeroCopySource(payload)); err != nil {
		log.Info("bcore.Msg", "#payload", len(payload), "type", bihsMsg.Type, "height", bihsMsg.Height, "view", bihsMsg.View, "qc", bihsMsg.Justify)
		return
	}

	log.Info("HandleP2pMsg", "#payload", len(payload), "msg hash", bihsMsg.Hash())

	select {
	case p.ch <- &bihsMsg:
	default:
		log.Warn("p2p msg dropped because channel is full")
	}

	return
}
