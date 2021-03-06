// Copyright (c) 2013 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcwire

import (
	"bytes"
	"io"
	"time"
)

// BlockVersion is the current latest supported block version.
const BlockVersion uint32 = 2

// Version 4 bytes + Timestamp 4 bytes + Bits 4 bytes + Nonce 4 bytes +
// TxnCount (varInt) + PrevBlock and MerkleRoot hashes.
const maxBlockHeaderPayload = 16 + maxVarIntPayload + (HashSize * 2)

// BlockHeader defines information about a block and is used in the bitcoin
// block (MsgBlock) and headers (MsgHeaders) messages.
type BlockHeader struct {
	// Version of the block.  This is not the same as the protocol version.
	Version uint32

	// Hash of the previous block in the block chain.
	PrevBlock ShaHash

	// Merkle tree reference to hash of all transactions for the block.
	MerkleRoot ShaHash

	// Time the block was created.  This is, unfortunately, encoded as a
	// uint32 on the wire and therefore is limited to 2106.
	Timestamp time.Time

	// Difficulty target for the block.
	Bits uint32

	// Nonce used to generate the block.
	Nonce uint32

	// Number of transactions in the block.  For the bitcoin headers
	// (MsgHeaders) message, this must be 0.  This is encoded as a variable
	// length integer on the wire.
	TxnCount uint64
}

// blockHashLen is a constant that represents how much of the block header is
// used when computing the block sha 0:blockHashLen
const blockHashLen = 80

// BlockSha computes the block identifier hash for the given block header.
func (h *BlockHeader) BlockSha(pver uint32) (sha ShaHash, err error) {
	var buf bytes.Buffer
	err = writeBlockHeader(&buf, pver, h)
	if err != nil {
		return
	}

	err = sha.SetBytes(DoubleSha256(buf.Bytes()[0:blockHashLen]))
	if err != nil {
		return
	}

	return
}

// NewBlockHeader returns a new BlockHeader using the provided previous block
// hash, merkle root hash, difficulty bits, and nonce used to generate the
// block with defaults for the remaining fields.
func NewBlockHeader(prevHash *ShaHash, merkleRootHash *ShaHash, bits uint32,
	nonce uint32) *BlockHeader {

	return &BlockHeader{
		Version:    BlockVersion,
		PrevBlock:  *prevHash,
		MerkleRoot: *merkleRootHash,
		Timestamp:  time.Now(),
		Bits:       bits,
		Nonce:      nonce,
		TxnCount:   0,
	}
}

// readBlockHeader reads a bitcoin block header from r.
func readBlockHeader(r io.Reader, pver uint32, bh *BlockHeader) error {
	var sec uint32
	err := readElements(r, &bh.Version, &bh.PrevBlock, &bh.MerkleRoot, &sec,
		&bh.Bits, &bh.Nonce)
	if err != nil {
		return err
	}
	bh.Timestamp = time.Unix(int64(sec), 0)

	count, err := readVarInt(r, pver)
	if err != nil {
		return err
	}
	bh.TxnCount = count

	return nil
}

// writeBlockHeader writes a bitcoin block header to w.
func writeBlockHeader(w io.Writer, pver uint32, bh *BlockHeader) error {
	sec := uint32(bh.Timestamp.Unix())
	err := writeElements(w, bh.Version, bh.PrevBlock, bh.MerkleRoot,
		sec, bh.Bits, bh.Nonce)
	if err != nil {
		return err
	}

	err = writeVarInt(w, pver, bh.TxnCount)
	if err != nil {
		return err
	}

	return nil
}
