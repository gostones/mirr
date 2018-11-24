package main

import (
	"fmt"
	"log"
	"sync"
)

// // Peer is
// type Peer struct {
// 	Address   string
// 	Rank      int // -1, 0, 1 ...
// 	timestamp int64
// }

// Neighborhood is
type Neighborhood struct {
	Peers map[string]*Peer
	min   int
	max   int

	sync.Mutex
}

// NewNeighborhood is
func NewNeighborhood() *Neighborhood {
	nb := &Neighborhood{
		Peers: make(map[string]*Peer, 15),
		min:   0,
		max:   1,
	}
	nb.Monitor()
	return nb
}

// GetPeers is
func (r *Neighborhood) GetPeers() []string {
	r.Lock()
	defer r.Unlock()

	addresses := make([]string, 0, len(r.Peers))
	for _, v := range r.Peers {
		if v.Rank > 0 {
			addresses = append(addresses, v.Addr)
		}
	}
	log.Printf("@@@@ addresses: %v\n", addresses)
	return addresses
}

// Monitor is
func (r *Neighborhood) Monitor() {
	job := func() {
		// clean up stale connections
		// for k, v := range r.Peers {
		// 	if v.Rank == 0 {
		// 		delete(r.Peers, k)
		// 	}
		// }

		cur := 0
		for _, v := range r.Peers {
			if v.Rank > 0 {
				cur++
			}
		}
		if cur >= r.max {
			log.Printf("@@@@ current count: %v max: %v, no new peers will be added\n", cur, r.max)
			return
		}

		//
		peers, err := p2pPeers()
		if err != nil {
			log.Printf("@@@@ get peers: %v\n", err)
		}

		cnt := len(peers)
		log.Printf("@@@@ get peers, count: %v\n", cnt)

		for i := 0; i < cnt; i++ {
			p := peers[i]
			id := p.Peer
			peer, found := r.Peers[id]

			log.Printf("@@@@ Peer ID: %v found: %v\n", id, found)

			if found {
				peer.Rank++
			} else {
				r.addPeer(id)
			}
		}
	}

	Every(1).Minutes().Run(job)
}

// addPeer is
func (r *Neighborhood) addPeer(id string) {
	r.Lock()
	defer r.Unlock()

	port := FreePort()
	addr := fmt.Sprintf("127.0.0.1:%v", port)
	// target := fmt.Sprintf("localhost:%v", config.ProxyPort)
	// go forward(addr, target)

	err := p2pForward(port, id)
	rank := -1
	if err == nil {
		ok := p2pIsValid(port)

		if ok {
			rank = 1
		}
	}
	log.Printf("@@@@ Add peer: %v addr: %v rank: %v err: %v\n", id, addr, rank, err)

	r.Peers[id] = &Peer{
		Addr:      addr,
		Rank:      rank,
		timestamp: CurrentTime(),
	}
}