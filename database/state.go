package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type State struct {
	Balances  map[Account]uint
	txMempool []Tx

	dbFile          *os.File
	latestBlockHash Hash
}

func NewStateFromDisk(dataDir string) (*State, error) {
	err := initDataDirIfNotExists(dataDir)
	if err != nil {
		return nil, err
	}

	gen, err := loadGenesis(getGenesisJsonFilePath(dataDir))
	if err != nil {
		return nil, err
	}

	balances := make(map[Account]uint)
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	f, err := os.OpenFile(getBlocksDbFilePath(dataDir), os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)

	state := &State{balances, make([]Tx, 0), f, Hash{}}

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		blockFsJson := scanner.Bytes()
		var blockFs BlockFS
		err = json.Unmarshal(blockFsJson, &blockFs)
		if err != nil {
			return nil, err
		}

		err = state.applyBlock(blockFs.Value)
		if err != nil {
			return nil, err
		}

		state.latestBlockHash = blockFs.Key
	}

	return state, nil
}

// func NewStateFromDisk() (*State, error) {
// 	cwd, err := os.Getwd()
// 	if err != nil {
// 		return nil, err
// 	}

// 	gen, err := loadGenesis(filepath.Join(cwd, "database", "genesis.json"))
// 	if err != nil {
// 		return nil, err
// 	}

// 	balances := make(map[Account]uint)
// 	for account, balance := range gen.Balances {
// 		balances[account] = balance
// 	}

// 	f, err := os.OpenFile(filepath.Join(cwd, "database", "block.db"), os.O_APPEND|os.O_RDWR, 0600)
// 	if err != nil {
// 		return nil, err
// 	}

// 	scanner := bufio.NewScanner(f)

// 	state := &State{balances, make([]Tx, 0), f, Hash{}}
// 	for scanner.Scan() {
// 		if err := scanner.Err(); err != nil {
// 			return nil, err
// 		}

// 		blockFsJson := scanner.Bytes()
// 		var blockFs BlockFS
// 		err = json.Unmarshal(blockFsJson, &blockFs)
// 		if err != nil {
// 			return nil, err
// 		}

// 		err = state.applyBlock(blockFs.Value)
// 		if err != nil {
// 			return nil, err
// 		}

// 		state.latestBlockHash = blockFs.Key
// 	}

// 	return state, nil
// }

func (s *State) LatestBlockHash() Hash {
	return s.latestBlockHash
}

// func (s *State) LatestSnapshot() Hash {
// 	return s.latestBlockHash
// }

func (s *State) AddBlock(b Block) error {
	for _, tx := range b.TXs {
		if err := s.AddTx(tx); err != nil {
			return err
		}
	}

	return nil
}

func (s *State) AddTx(tx Tx) error {
	if err := s.apply(tx); err != nil {
		return err
	}

	s.txMempool = append(s.txMempool, tx)

	return nil
}

func (s *State) Persist() (Hash, error) {
	// Create a new Block with ONLY the new TXs
	block := NewBlock(s.latestBlockHash, uint64(time.Now().Unix()), s.txMempool)
	blockHash, err := block.Hash()

	_ = blockHash
	if err != nil {
		return Hash{}, err
	}

	blockFS := BlockFS{blockHash, block}

	// Encode it into a Json string
	blockFsJson, err := json.Marshal(blockFS)
	if err != nil {
		return Hash{}, err
	}

	fmt.Printf("Persisting new Block to disk:\n")
	fmt.Printf("\t%s\n", blockFsJson)

	//Write it to the DB file on a new line
	if _, err = s.dbFile.Write(append(blockFsJson, '\n')); err != nil {
		return Hash{}, err
	}

	s.latestBlockHash = blockHash

	s.txMempool = []Tx{} // reset the mempool

	return s.latestBlockHash, nil
}

func (s *State) Close() {
	s.dbFile.Close()
}

func (s *State) apply(tx Tx) error {
	if tx.IsReward() {
		s.Balances[tx.To] += tx.Value
		return nil
	}

	if s.Balances[tx.From] < tx.Value {
		return fmt.Errorf("insufficient balance")
	}

	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value

	return nil
}

func (s *State) applyBlock(b Block) error {
	for _, tx := range b.TXs {
		if err := s.apply(tx); err != nil {
			return err
		}
	}

	return nil
}
