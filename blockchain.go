// implement a simple p2p blockchain which use dpos algorithm.
// this file just for a simple block generate and validate.

package dpos

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// BlockChain slice to storage Block
var BlockChain []Block

type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
	validator string
}

func CaculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// CaculateBlockHash 
func CaculateBlockHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
	return CaculateHash(record)
}

// GenerateBlock 
func GenerateBlock(oldBlock Block, BPM int, address string) (Block, error) {
	var newBlock Block
	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.BPM = BPM
	newBlock.Timestamp = t.String()
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = CaculateBlockHash(newBlock)
	newBlock.validator = address

	return newBlock, nil
}

// IsBlockValid 
func IsBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if CaculateBlockHash(newBlock) != newBlock.Hash {
		return false
	}
	return true
}
