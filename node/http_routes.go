package node

import (
	"net/http"

	"github.com/web3coach/the-blockchain-bar/database"
)

type ErrRes struct {
	Error string `json:"error"`
}

type BalancesRes struct {
	Hash     database.Hash             `json:"block_hash"`
	Balances map[database.Account]uint `json:"balances"`
}

type TxAddReq struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint   `json:"value"`
	Data  string `json:"data"`
}

type TxAddRes struct {
	Hash database.Hash `json:"block_hash"`
}

type StatusRes struct {
	Hash       database.Hash `json:"block_hash"`
	Number     uint64        `json:"block_number"`
	KnownPeers []PeerNode    `json:"peers_known"` // tell me your peers
}

type SyncRes struct {
	Blocks []database.Block `json:"blocks"`
}

type AddPeerRes struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func statusHandler(w http.ResponseWriter, r *http.Request, n *Node) {
	res := StatusRes{
		Hash:       n.state.LatestBlockHash(),
		Number:     n.state.LatestBlock().Header.Number,
		KnownPeers: n.knownPeers,
	}
	writeRes(w, res)
}
