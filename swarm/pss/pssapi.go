package pss

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm/network"
)

// PssAPI is the RPC API module for Pss
type PssAPI struct {
	*Pss
}

// NewPssAPI constructs a PssAPI instance
func NewPssAPI(ps *Pss) *PssAPI {
	return &PssAPI{Pss: ps}
}

// temporary for access to overlay while faking kademlia healthy routines
func (pssapi *PssAPI) GetForwarder(addr []byte) (fwd struct {
	Addr  []byte
	Count int
}) {
	pssapi.Pss.Overlay.EachConn(addr, 255, func(op network.OverlayConn, po int, isproxbin bool) bool {
		//pssapi.Pss.Overlay.EachLivePeer(addr, 255, func(p network.Peer, po int, isprox bool) bool {
		if bytes.Equal(fwd.Addr, []byte{}) {
			fwd.Addr = op.Address()
		}
		fwd.Count++
		return true
	})
	return
}

// NewMsg API endpoint creates an RPC subscription
func (pssapi *PssAPI) NewMsg(ctx context.Context, topic PssTopic) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return nil, fmt.Errorf("Subscribe not supported")
	}

	psssub := notifier.CreateSubscription()
	handler := func(msg []byte, p *p2p.Peer, from []byte) error {
		apimsg := &PssAPIMsg{
			Msg:  msg,
			Addr: from,
		}
		if err := notifier.Notify(psssub.ID, apimsg); err != nil {
			log.Warn(fmt.Sprintf("notification on pss sub topic %v rpc (sub %v) msg %v failed!", topic, psssub.ID, msg))
		}
		return nil
	}
	deregf := pssapi.Pss.Register(&topic, handler)

	go func() {
		defer deregf()
		//defer psssub.Unsubscribe()
		select {
		case err := <-psssub.Err():
			log.Warn(fmt.Sprintf("caught subscription error in pss sub topic %x: %v", topic, err))
		case <-notifier.Closed():
			log.Warn(fmt.Sprintf("rpc sub notifier closed"))
		}
	}()

	return psssub, nil
}

// SendRaw sends the message (serialized into byte slice) to a peer with topic
func (pssapi *PssAPI) SendRaw(topic PssTopic, msg PssAPIMsg) error {
	err := pssapi.Pss.Send(msg.Addr, topic, msg.Msg)
	if err != nil {
		return fmt.Errorf("send error: %v", err)
	}
	return fmt.Errorf("ok sent")
}

// BaseAddr gets our own overlayaddress
func (pssapi *PssAPI) BaseAddr() ([]byte, error) {
	log.Warn("inside baseaddr")
	return pssapi.Pss.BaseAddr(), nil
}
