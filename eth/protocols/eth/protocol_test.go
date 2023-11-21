// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package eth

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/prysmaticlabs/prysm/crypto/bls/blst"
)

// Tests that the custom union field encoder and decoder works correctly.
func TestGetBlockHeadersDataEncodeDecode(t *testing.T) {
	// Create a "random" hash for testing
	var hash common.Hash
	for i := range hash {
		hash[i] = byte(i)
	}
	// Assemble some table driven tests
	tests := []struct {
		packet *GetBlockHeadersPacket
		fail   bool
	}{
		// Providing the origin as either a hash or a number should both work
		{fail: false, packet: &GetBlockHeadersPacket{Origin: HashOrNumber{Number: 314}}},
		{fail: false, packet: &GetBlockHeadersPacket{Origin: HashOrNumber{Hash: hash}}},

		// Providing arbitrary query field should also work
		{fail: false, packet: &GetBlockHeadersPacket{Origin: HashOrNumber{Number: 314}, Amount: 314, Skip: 1, Reverse: true}},
		{fail: false, packet: &GetBlockHeadersPacket{Origin: HashOrNumber{Hash: hash}, Amount: 314, Skip: 1, Reverse: true}},

		// Providing both the origin hash and origin number must fail
		{fail: true, packet: &GetBlockHeadersPacket{Origin: HashOrNumber{Hash: hash, Number: 314}}},
	}
	// Iterate over each of the tests and try to encode and then decode
	for i, tt := range tests {
		bytes, err := rlp.EncodeToBytes(tt.packet)
		if err != nil && !tt.fail {
			t.Fatalf("test %d: failed to encode packet: %v", i, err)
		} else if err == nil && tt.fail {
			t.Fatalf("test %d: encode should have failed", i)
		}
		if !tt.fail {
			packet := new(GetBlockHeadersPacket)
			if err := rlp.DecodeBytes(bytes, packet); err != nil {
				t.Fatalf("test %d: failed to decode packet: %v", i, err)
			}
			if packet.Origin.Hash != tt.packet.Origin.Hash || packet.Origin.Number != tt.packet.Origin.Number || packet.Amount != tt.packet.Amount ||
				packet.Skip != tt.packet.Skip || packet.Reverse != tt.packet.Reverse {
				t.Fatalf("test %d: encode decode mismatch: have %+v, want %+v", i, packet, tt.packet)
			}
		}
	}
}

// TestEth66EmptyMessages tests encoding of empty eth66 messages
func TestEth66EmptyMessages(t *testing.T) {
	// All empty messages encodes to the same format
	want := common.FromHex("c4820457c0")

	for i, msg := range []interface{}{
		// Headers
		GetBlockHeadersPacket66{1111, nil},
		BlockHeadersPacket66{1111, nil},
		// Bodies
		GetBlockBodiesPacket66{1111, nil},
		BlockBodiesPacket66{1111, nil},
		BlockBodiesRLPPacket66{1111, nil},
		// Node data
		GetNodeDataPacket66{1111, nil},
		NodeDataPacket66{1111, nil},
		// Receipts
		GetReceiptsPacket66{1111, nil},
		ReceiptsPacket66{1111, nil},
		// Transactions
		GetPooledTransactionsPacket66{1111, nil},
		PooledTransactionsPacket66{1111, nil},
		PooledTransactionsRLPPacket66{1111, nil},

		// Headers
		BlockHeadersPacket66{1111, BlockHeadersPacket([]*types.Header{})},
		// Bodies
		GetBlockBodiesPacket66{1111, GetBlockBodiesPacket([]common.Hash{})},
		BlockBodiesPacket66{1111, BlockBodiesPacket([]*BlockBody{})},
		BlockBodiesRLPPacket66{1111, BlockBodiesRLPPacket([]rlp.RawValue{})},
		// Node data
		GetNodeDataPacket66{1111, GetNodeDataPacket([]common.Hash{})},
		NodeDataPacket66{1111, NodeDataPacket([][]byte{})},
		// Receipts
		GetReceiptsPacket66{1111, GetReceiptsPacket([]common.Hash{})},
		ReceiptsPacket66{1111, ReceiptsPacket([][]*types.Receipt{})},
		// Transactions
		GetPooledTransactionsPacket66{1111, GetPooledTransactionsPacket([]common.Hash{})},
		PooledTransactionsPacket66{1111, PooledTransactionsPacket([]*types.Transaction{})},
		PooledTransactionsRLPPacket66{1111, PooledTransactionsRLPPacket([]rlp.RawValue{})},
	} {
		if have, _ := rlp.EncodeToBytes(msg); !bytes.Equal(have, want) {
			t.Errorf("test %d, type %T, have\n\t%x\nwant\n\t%x", i, msg, have, want)
		}
	}

}

// TestEth66Messages tests the encoding of all redefined eth66 messages
func TestEth66Messages(t *testing.T) {

	// Some basic structs used during testing
	var (
		header       *types.Header
		blockBody    *BlockBody
		blockBodyRlp rlp.RawValue
		txs          []*types.Transaction
		txRlps       []rlp.RawValue
		hashes       []common.Hash
		receipts     []*types.Receipt
		receiptsRlp  rlp.RawValue

		err error
	)
	header = &types.Header{
		Difficulty: big.NewInt(2222),
		Number:     big.NewInt(3333),
		GasLimit:   4444,
		GasUsed:    5555,
		Time:       6666,
		Extra:      []byte{0x77, 0x88},
	}
	// Init the transactions, taken from a different test
	{
		for _, hexrlp := range []string{
			"f867088504a817c8088302e2489435353535353535353535353535353535353535358202008025a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c12a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c10",
			"f867098504a817c809830334509435353535353535353535353535353535353535358202d98025a052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afba052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afb",
		} {
			var tx *types.Transaction
			rlpdata := common.FromHex(hexrlp)
			if err := rlp.DecodeBytes(rlpdata, &tx); err != nil {
				t.Fatal(err)
			}
			txs = append(txs, tx)
			txRlps = append(txRlps, rlpdata)
		}
	}
	// init the block body data, both object and rlp form
	blockBody = &BlockBody{
		Transactions: txs,
		Uncles:       []*types.Header{header},
	}
	blockBodyRlp, err = rlp.EncodeToBytes(blockBody)
	if err != nil {
		t.Fatal(err)
	}

	hashes = []common.Hash{
		common.HexToHash("deadc0de"),
		common.HexToHash("feedbeef"),
	}
	byteSlices := [][]byte{
		common.FromHex("deadc0de"),
		common.FromHex("feedbeef"),
	}
	// init the receipts
	{
		receipts = []*types.Receipt{
			{
				Status:            types.ReceiptStatusFailed,
				CumulativeGasUsed: 1,
				Logs: []*types.Log{
					{
						Address: common.BytesToAddress([]byte{0x11}),
						Topics:  []common.Hash{common.HexToHash("dead"), common.HexToHash("beef")},
						Data:    []byte{0x01, 0x00, 0xff},
					},
				},
				TxHash:          hashes[0],
				ContractAddress: common.BytesToAddress([]byte{0x01, 0x11, 0x11}),
				GasUsed:         111111,
			},
		}
		rlpData, err := rlp.EncodeToBytes(receipts)
		if err != nil {
			t.Fatal(err)
		}
		receiptsRlp = rlpData
	}

	for i, tc := range []struct {
		message interface{}
		want    []byte
	}{
		{
			GetBlockHeadersPacket66{1111, &GetBlockHeadersPacket{HashOrNumber{hashes[0], 0}, 5, 5, false}},
			common.FromHex("e8820457e4a000000000000000000000000000000000000000000000000000000000deadc0de050580"),
		},
		{
			GetBlockHeadersPacket66{1111, &GetBlockHeadersPacket{HashOrNumber{common.Hash{}, 9999}, 5, 5, false}},
			common.FromHex("ca820457c682270f050580"),
		},
		{
			BlockHeadersPacket66{1111, BlockHeadersPacket{header}},
			common.FromHex("f90202820457f901fcf901f9a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000940000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008208ae820d0582115c8215b3821a0a827788a00000000000000000000000000000000000000000000000000000000000000000880000000000000000"),
		},
		{
			GetBlockBodiesPacket66{1111, GetBlockBodiesPacket(hashes)},
			common.FromHex("f847820457f842a000000000000000000000000000000000000000000000000000000000deadc0dea000000000000000000000000000000000000000000000000000000000feedbeef"),
		},
		{
			BlockBodiesPacket66{1111, BlockBodiesPacket([]*BlockBody{blockBody})},
			common.FromHex("f902dc820457f902d6f902d3f8d2f867088504a817c8088302e2489435353535353535353535353535353535353535358202008025a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c12a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c10f867098504a817c809830334509435353535353535353535353535353535353535358202d98025a052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afba052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afbf901fcf901f9a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000940000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008208ae820d0582115c8215b3821a0a827788a00000000000000000000000000000000000000000000000000000000000000000880000000000000000"),
		},
		{ // Identical to non-rlp-shortcut version
			BlockBodiesRLPPacket66{1111, BlockBodiesRLPPacket([]rlp.RawValue{blockBodyRlp})},
			common.FromHex("f902dc820457f902d6f902d3f8d2f867088504a817c8088302e2489435353535353535353535353535353535353535358202008025a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c12a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c10f867098504a817c809830334509435353535353535353535353535353535353535358202d98025a052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afba052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afbf901fcf901f9a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000940000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008208ae820d0582115c8215b3821a0a827788a00000000000000000000000000000000000000000000000000000000000000000880000000000000000"),
		},
		{
			GetNodeDataPacket66{1111, GetNodeDataPacket(hashes)},
			common.FromHex("f847820457f842a000000000000000000000000000000000000000000000000000000000deadc0dea000000000000000000000000000000000000000000000000000000000feedbeef"),
		},
		{
			NodeDataPacket66{1111, NodeDataPacket(byteSlices)},
			common.FromHex("ce820457ca84deadc0de84feedbeef"),
		},
		{
			GetReceiptsPacket66{1111, GetReceiptsPacket(hashes)},
			common.FromHex("f847820457f842a000000000000000000000000000000000000000000000000000000000deadc0dea000000000000000000000000000000000000000000000000000000000feedbeef"),
		},
		{
			ReceiptsPacket66{1111, ReceiptsPacket([][]*types.Receipt{receipts})},
			common.FromHex("f90172820457f9016cf90169f901668001b9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f85ff85d940000000000000000000000000000000000000011f842a0000000000000000000000000000000000000000000000000000000000000deada0000000000000000000000000000000000000000000000000000000000000beef830100ff"),
		},
		{
			ReceiptsRLPPacket66{1111, ReceiptsRLPPacket([]rlp.RawValue{receiptsRlp})},
			common.FromHex("f90172820457f9016cf90169f901668001b9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f85ff85d940000000000000000000000000000000000000011f842a0000000000000000000000000000000000000000000000000000000000000deada0000000000000000000000000000000000000000000000000000000000000beef830100ff"),
		},
		{
			GetPooledTransactionsPacket66{1111, GetPooledTransactionsPacket(hashes)},
			common.FromHex("f847820457f842a000000000000000000000000000000000000000000000000000000000deadc0dea000000000000000000000000000000000000000000000000000000000feedbeef"),
		},
		{
			PooledTransactionsPacket66{1111, PooledTransactionsPacket(txs)},
			common.FromHex("f8d7820457f8d2f867088504a817c8088302e2489435353535353535353535353535353535353535358202008025a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c12a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c10f867098504a817c809830334509435353535353535353535353535353535353535358202d98025a052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afba052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afb"),
		},
		{
			PooledTransactionsRLPPacket66{1111, PooledTransactionsRLPPacket(txRlps)},
			common.FromHex("f8d7820457f8d2f867088504a817c8088302e2489435353535353535353535353535353535353535358202008025a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c12a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c10f867098504a817c809830334509435353535353535353535353535353535353535358202d98025a052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afba052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afb"),
		},
	} {
		if have, _ := rlp.EncodeToBytes(tc.message); !bytes.Equal(have, tc.want) {
			t.Errorf("test %d, type %T, have\n\t%x\nwant\n\t%x", i, tc.message, have, tc.want)
		}
	}
}

// TestEth68Messages tests the encoding of newly defined eth68 messages
func TestEth68Messages(t *testing.T) {

	// Some basic structs used during testing
	var (
		BLSPrivateKey = "4cf9fc19af38d1bbaf85b3639502f9eef4bc90c196fe36cc0252abf51551c8bd"

		votesSet []*types.VoteEnvelope
	)

	// Init the vote data, taken from a local node
	secretKey, _ := blst.SecretKeyFromBytes(common.Hex2Bytes(BLSPrivateKey))
	{
		for _, voteData := range []types.VoteData{
			{
				SourceNumber: 0,
				SourceHash:   common.HexToHash("0x6d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34"),
				TargetNumber: 1,
				TargetHash:   common.HexToHash("0xd0bc67b50915467ada963c35ee00950f664788e47da8139d8c178653171034f1"),
			},
			{
				SourceNumber: 0,
				SourceHash:   common.HexToHash("0x6d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34"),
				TargetNumber: 2,
				TargetHash:   common.HexToHash("0xc2d18d5a59d65da573f70c4d30448482418894e018b0d189db24ea4fd02d7aa1"),
			},
			{
				SourceNumber: 0,
				SourceHash:   common.HexToHash("0x6d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34"),
				TargetNumber: 4,
				TargetHash:   common.HexToHash("0xbd1bdaf8a8f5c00c464df2856a9e2ef23b8dcc906e6490d3cd295ebb5eb124c3"),
			},
			{
				SourceNumber: 0,
				SourceHash:   common.HexToHash("0x6d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34"),
				TargetNumber: 8,
				TargetHash:   common.HexToHash("0x3073782ecabb5ef0673e95962273482347a2c7b30a0a7124c664443d0a43f1e1"),
			},
			{
				SourceNumber: 0,
				SourceHash:   common.HexToHash("0x6d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34"),
				TargetNumber: 16,
				TargetHash:   common.HexToHash("0xc119833266327fd7e0cd929c6a847ae7d1689df5066dfdde2e52f51c0ecbcc3f"),
			},
			{
				SourceNumber: 0,
				SourceHash:   common.HexToHash("0x6d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34"),
				TargetNumber: 32,
				TargetHash:   common.HexToHash("0x3b5650bcb98381e463871a15a3f601cdc26843d76f4d3461333d7feae38a1786"),
			},
			{
				SourceNumber: 0,
				SourceHash:   common.HexToHash("0x6d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34"),
				TargetNumber: 64,
				TargetHash:   common.HexToHash("0x5e38b4d98904178d60d58f5bc1687b0c7df114a51f2007d3ee3e6e732539f130"),
			},
			{
				SourceNumber: 0,
				SourceHash:   common.HexToHash("0x6d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34"),
				TargetNumber: 128,
				TargetHash:   common.HexToHash("0xa4a64a7d511d3ff6982b5a79e9a485508477b98996c570a220f9daea0c7682f8"),
			},
			{
				SourceNumber: 0,
				SourceHash:   common.HexToHash("0x6d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34"),
				TargetNumber: 256,
				TargetHash:   common.HexToHash("0xd313672c2653fc13e75a9dafdcee93f444caf2cffb04585d3e306fd15418b7e2"),
			},
			{
				SourceNumber: 0,
				SourceHash:   common.HexToHash("0x6d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34"),
				TargetNumber: 512,
				TargetHash:   common.HexToHash("0x3c5fe2e5439ca7a7f1a3de7d5c0914c37261451c87654397dd45f207109839ae"),
			},
			{
				SourceNumber: 0,
				SourceHash:   common.HexToHash("0x6d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34"),
				TargetNumber: 1024,
				TargetHash:   common.HexToHash("0x088eeeb07acff0db3ae2585195e9fd23bdf54b55077cab87d1632b08dd2c043b"),
			},
		} {
			votes := new(types.VoteEnvelope)
			voteAddress := new(types.BLSPublicKey)
			signature := new(types.BLSSignature)
			copy(voteAddress[:], secretKey.PublicKey().Marshal()[:])
			copy(signature[:], secretKey.Sign(voteData.Hash().Bytes()).Marshal()[:])
			votes.VoteAddress = *voteAddress
			votes.Signature = *signature
			votes.Data = &voteData
			votesSet = append(votesSet, votes)
		}
	}

	for i, tc := range []struct {
		message interface{}
		want    []byte
	}{
		{
			VotesPacket{votesSet},
			common.FromHex("f90982f9097ff8dbb0b32d4d46a7127dcc865f0d30f2ee3dcd5983b686f4e3a9202afc8b608652001c9938906ae1ff1417486096e32511f1bcb86091f8a39f99a0b3632d248e635ab0e7b5ff68071fa4def5c2df9f249db7d393c0ba89eb28a65f2a6ba836baddb961b9c312c70a87d130edf944b340649218335c91078cce808da75ff69f673bab3ecdf068c33b1ab147c54298056b19e9cc625df84680a06d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34820400a0088eeeb07acff0db3ae2585195e9fd23bdf54b55077cab87d1632b08dd2c043bf8dbb0b32d4d46a7127dcc865f0d30f2ee3dcd5983b686f4e3a9202afc8b608652001c9938906ae1ff1417486096e32511f1bcb860a56cd257f9a4b4830c9bfadaa751c7b1d4e9c6899127c145987a55a7cfa0d1b7d114cb2523ea4e2efee0326cfc5a1cd912eaf7f0c4c0be3193677284533f1709fbd75471a9fb22aea358cdbf2e900628d7c504ce7245e8af6fdd1039dfa3c0bdf84680a06d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34820400a0088eeeb07acff0db3ae2585195e9fd23bdf54b55077cab87d1632b08dd2c043bf8dbb0b32d4d46a7127dcc865f0d30f2ee3dcd5983b686f4e3a9202afc8b608652001c9938906ae1ff1417486096e32511f1bcb860893f8aff7fc523a7aff006aaba71fbde5f1eee1f4683d405892ffb9ab9282a5dae01054210ff6873ee76f86b9afdef2e128932b26696e3f7e1de7fe7d3fdd1c41273912ff5d1002cba176dbf84e1fe2edb60b114129b89e1329a03f7d9843d04f84680a06d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34820400a0088eeeb07acff0db3ae2585195e9fd23bdf54b55077cab87d1632b08dd2c043bf8dbb0b32d4d46a7127dcc865f0d30f2ee3dcd5983b686f4e3a9202afc8b608652001c9938906ae1ff1417486096e32511f1bcb860b3585bf55f1e0d8bc0f544a386e6fc4ec37de52330f69b450f579050acda6279a8a38172ed2f01dfdb57cf7dd2a314970aa8a3168234cbd29adfc6a0efd080f57d7e195dafbf5b6db087e8b943aa634f797f8f6d4e5bf04681d7ce2218e59465f84680a06d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34820400a0088eeeb07acff0db3ae2585195e9fd23bdf54b55077cab87d1632b08dd2c043bf8dbb0b32d4d46a7127dcc865f0d30f2ee3dcd5983b686f4e3a9202afc8b608652001c9938906ae1ff1417486096e32511f1bcb8609366e823b456049cd10ed1aa8f02a58ce2fa4caea7e8c776d6aec9c42f4263b40b0f2d76cc55a598b378381f32ef33520d47e28707027c25e38eb971cddb379e0ded5e814ce70108d65855084a11484fd08447520b7ce79ac1e680020b243747f84680a06d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34820400a0088eeeb07acff0db3ae2585195e9fd23bdf54b55077cab87d1632b08dd2c043bf8dbb0b32d4d46a7127dcc865f0d30f2ee3dcd5983b686f4e3a9202afc8b608652001c9938906ae1ff1417486096e32511f1bcb860aafd383c9537d750358ea077d45476cf6c1541e717c690ebe5dc5442c2af732fba17b45c60b2c49de94f5121f318b6ae021c56ae06588c6552f1d5b87a166cb8050f287b528b20556033603f6a6649ccec4792c86ae5f6353cf6b7527ac40127f84680a06d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34820400a0088eeeb07acff0db3ae2585195e9fd23bdf54b55077cab87d1632b08dd2c043bf8dbb0b32d4d46a7127dcc865f0d30f2ee3dcd5983b686f4e3a9202afc8b608652001c9938906ae1ff1417486096e32511f1bcb86090d9dc467a64fe7852b2e0117806409a8f12a036849d69396282088f8d86adb3adcd46b1fde51b4630a6b64c2f8652f30a46609c49b33f50c9f4641e30900ee420f9b81b2ad59a2376dcf4e065ecf832fbf738ad5b73becd2f7add27e6c14d5ff84680a06d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34820400a0088eeeb07acff0db3ae2585195e9fd23bdf54b55077cab87d1632b08dd2c043bf8dbb0b32d4d46a7127dcc865f0d30f2ee3dcd5983b686f4e3a9202afc8b608652001c9938906ae1ff1417486096e32511f1bcb8608f7d6bc28626dc143208aaa7b97146510f01b1a108dead65f8fddf0ec07835bca91081f9e759656d52dd7d4adaac14220c8c62aa1dd09151fe8000ce4347b100ac496593058ae11b40c74b3049d38076d07301c9dc9586baf93d9c81b4e5d424f84680a06d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34820400a0088eeeb07acff0db3ae2585195e9fd23bdf54b55077cab87d1632b08dd2c043bf8dbb0b32d4d46a7127dcc865f0d30f2ee3dcd5983b686f4e3a9202afc8b608652001c9938906ae1ff1417486096e32511f1bcb860b6c17077217baa5930fb04e14b6ba7203b3c854e8f1363b48ad753d96db1b4ffed36d60d8b67e86f7f76504f0abefff80ed1e4f11ff192dbfc26a0f059f7b9f66f9e883fef208cc3f58c7ce49d8e854cf8a0e48c59c7407ebfe946cfd62bf3bef84680a06d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34820400a0088eeeb07acff0db3ae2585195e9fd23bdf54b55077cab87d1632b08dd2c043bf8dbb0b32d4d46a7127dcc865f0d30f2ee3dcd5983b686f4e3a9202afc8b608652001c9938906ae1ff1417486096e32511f1bcb860979b1d101e51731749c72fb160dd1245d69ebd6ca81c0519464d3bca9ec3db493cf4b45ebbb7f60fbd12f0705bd788000558bdedc335cedac2100169965b2794fae8a386b2e9ece86ea6952fadeb8501d9aad00e091713cc06d30c5885c3ecf0f84680a06d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34820400a0088eeeb07acff0db3ae2585195e9fd23bdf54b55077cab87d1632b08dd2c043bf8dbb0b32d4d46a7127dcc865f0d30f2ee3dcd5983b686f4e3a9202afc8b608652001c9938906ae1ff1417486096e32511f1bcb8608d035b04d8ef6c13117acc1ed9d0586a141f123494f21eeaaead5dd9f623933541b293eef403d2f3e8ede84f9dfe3dc10cbd3fa6bdf3e977dcf2d18a4dca84f8bd9b24fca8e7de4180b9aa6208ad6e756b1c81e98afc8e6994824b5c076857f8f84680a06d3c66c5357ec91d5c43af47e234a939b22557cbb552dc45bebbceeed90fbe34820400a0088eeeb07acff0db3ae2585195e9fd23bdf54b55077cab87d1632b08dd2c043b"),
		},
	} {
		if have, _ := rlp.EncodeToBytes(tc.message); !bytes.Equal(have, tc.want) {
			t.Errorf("test %d, type %T, have\n\t%x\nwant\n\t%x", i, tc.message, have, tc.want)
		}
	}
}
