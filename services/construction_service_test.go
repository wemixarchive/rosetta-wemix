// Copyright 2020 Coinbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/wemixarchive/rosetta-wemix/configuration"
	mocks "github.com/wemixarchive/rosetta-wemix/mocks/services"
	"github.com/wemixarchive/rosetta-wemix/wemix"
	// "github.com/metadium/rosetta-metadium/params"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func forceHexDecode(t *testing.T, s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("could not decode hex %s", s)
	}

	return b
}

func forceMarshalMap(t *testing.T, i interface{}) map[string]interface{} {
	m, err := marshalJSONMap(i)
	if err != nil {
		t.Fatalf("could not marshal map %s", types.PrintStruct(i))
	}

	return m
}
func log(t *testing.T, label string, o interface{}) {
	tmp, _ := json.Marshal(o)
	t.Log(label, string(tmp))
	//fmt.Println(label, string(tmp))
}

func TestConstructionService(t *testing.T) {

	networkIdentifier = &types.NetworkIdentifier{
		Network:    wemix.TestnetNetwork,
		Blockchain: wemix.Blockchain,
	}

	cfg := &configuration.Configuration{
		Mode:    configuration.Online,
		Network: networkIdentifier,
		Params:  params.WemixTestnetChainConfig,
	}

	mockClient := &mocks.Client{}
	servicer := NewConstructionAPIService(cfg, mockClient)
	ctx := context.Background()

	// Test Derive
	publicKey := &types.PublicKey{
		Bytes: forceHexDecode(
			t,
			//"03d3d3358e7f69cbe45bde38d7d6f24660c7eeeaee5c5590cfab985c8839b21fd5", //compress
			//			"04d3d3358e7f69cbe45bde38d7d6f24660c7eeeaee5c5590cfab985c8839b21fd5c8dfcc91f42b1d80ee495887ba70616e8f15ad1d137eba4e2e9c4cc340166bb3", //uncompress
			"036d9038945ff8f4669201ba1e806c9a46a5034a578e4d52c03152198538039294",
			// "046d9038945ff8f4669201ba1e806c9a46a5034a578e4d52c031521985380392944efd6c702504d9130573bb939f5c124af95d38168546cc7207a7e0baf14172ff",
		),
		CurveType: types.Secp256k1,
	}
	deriveResponse, err := servicer.ConstructionDerive(ctx, &types.ConstructionDeriveRequest{
		NetworkIdentifier: networkIdentifier,
		PublicKey:         publicKey,
	})

	log(t, "deriveResponse", deriveResponse)
	assert.Nil(t, err)
	assert.Equal(t, &types.ConstructionDeriveResponse{
		AccountIdentifier: &types.AccountIdentifier{
			//			Address: "0xe3a5B4d7f79d64088C8d4ef153A7DDe2B2d47309",
			Address: "0xbe862AD9AbFe6f22BCb087716c7D89a26051f74C",
		},
	}, deriveResponse)

	// Test Preprocess
	intent := `[{"operation_identifier":{"index":0},"type":"CALL","account":{"address":"0xb22694a52EA2a9564001aF4AA61ecD9672E0D26b"},"amount":{"value":"-42894881044106498","currency":{"symbol":"WEMIX","decimals":18}}},{"operation_identifier":{"index":1},"type":"CALL","account":{"address":"0x57B414a0332B5CaB885a451c2a28a07d1e9b8a8d"},"amount":{"value":"42894881044106498","currency":{"symbol":"WEMIX","decimals":18}}}]` // nolint
	var ops []*types.Operation
	assert.NoError(t, json.Unmarshal([]byte(intent), &ops))

	log(t, "ops", ops)
	preprocessResponse, err := servicer.ConstructionPreprocess(
		ctx,
		&types.ConstructionPreprocessRequest{
			NetworkIdentifier: networkIdentifier,
			Operations:        ops,
		},
	)
	log(t, "preprocessResponse", preprocessResponse)
	assert.Nil(t, err)
	optionsRaw := `{"from":"0xb22694a52EA2a9564001aF4AA61ecD9672E0D26b"}`
	var options options
	assert.NoError(t, json.Unmarshal([]byte(optionsRaw), &options))
	log(t, "options", options)
	assert.Equal(t, &types.ConstructionPreprocessResponse{
		Options: forceMarshalMap(t, options),
	}, preprocessResponse)

	// Test Metadata
	metadata := &metadata{
		GasPrice: big.NewInt(80000000000),
		Nonce:    0,
	}

	mockClient.On(
		"SuggestGasPrice",
		ctx,
	).Return(
		big.NewInt(80000000000), //1000000000
		nil,
	).Once()
	mockClient.On(
		"PendingNonceAt",
		ctx,
		common.HexToAddress("0xb22694a52EA2a9564001aF4AA61ecD9672E0D26b"),
	).Return(
		uint64(0),
		nil,
	).Once()
	metadataResponse, err := servicer.ConstructionMetadata(ctx, &types.ConstructionMetadataRequest{
		NetworkIdentifier: networkIdentifier,
		Options:           forceMarshalMap(t, options),
	})
	assert.Nil(t, err)
	assert.Equal(t, &types.ConstructionMetadataResponse{
		Metadata: forceMarshalMap(t, metadata),
		SuggestedFee: []*types.Amount{
			{
				Value:    "1680000000000000", //"21000000000000",
				Currency: wemix.Currency,
			},
		},
	}, metadataResponse)

	// Test Payloads  80000000000
	unsignedRaw := `{"from":"0xb22694a52EA2a9564001aF4AA61ecD9672E0D26b","to":"0x57B414a0332B5CaB885a451c2a28a07d1e9b8a8d","value":"0x9864aac3510d02","data":"0x","nonce":"0x0","gas_price":"0x12a05f2000","gas":"0x5208","chain_id":"0x458"}` // nolint
	payloadsResponse, err := servicer.ConstructionPayloads(ctx, &types.ConstructionPayloadsRequest{
		NetworkIdentifier: networkIdentifier,
		Operations:        ops,
		Metadata:          forceMarshalMap(t, metadata),
	})
	log(t, "payloadsResponse: ", payloadsResponse)
	assert.Nil(t, err)
	payloadsRaw := `[{"address":"0xbe862AD9AbFe6f22BCb087716c7D89a26051f74C","hex_bytes":"996836219400142b587d5ad87b1f70d25a2497fd6ac431509ed90c48df8b2a9f","account_identifier":{"address":"0xb22694a52EA2a9564001aF4AA61ecD9672E0D26b"},"signature_type":"ecdsa_recovery"}]` // nolint
	var payloads []*types.SigningPayload
	assert.NoError(t, json.Unmarshal([]byte(payloadsRaw), &payloads))
	log(t, "payload: ", payloads)
	assert.Equal(t, &types.ConstructionPayloadsResponse{
		UnsignedTransaction: unsignedRaw,
		Payloads:            payloads,
	}, payloadsResponse)

	// Test Parse Unsigned
	parseOpsRaw := `[{"operation_identifier":{"index":0},"type":"CALL","account":{"address":"0xb22694a52EA2a9564001aF4AA61ecD9672E0D26b"},"amount":{"value":"-42894881044106498","currency":{"symbol":"WEMIX","decimals":18}}},{"operation_identifier":{"index":1},"related_operations":[{"index":0}],"type":"CALL","account":{"address":"0x57B414a0332B5CaB885a451c2a28a07d1e9b8a8d"},"amount":{"value":"42894881044106498","currency":{"symbol":"WEMIX","decimals":18}}}]` // nolint
	var parseOps []*types.Operation
	assert.NoError(t, json.Unmarshal([]byte(parseOpsRaw), &parseOps))
	parseUnsignedResponse, err := servicer.ConstructionParse(ctx, &types.ConstructionParseRequest{
		NetworkIdentifier: networkIdentifier,
		Signed:            false,
		Transaction:       unsignedRaw,
	})
	assert.Nil(t, err)
	parseMetadata := &parseMetadata{
		Nonce:    metadata.Nonce,
		GasPrice: metadata.GasPrice,
		ChainID:  big.NewInt(1112),
	}
	assert.Equal(t, &types.ConstructionParseResponse{
		Operations:               parseOps,
		AccountIdentifierSigners: []*types.AccountIdentifier{},
		Metadata:                 forceMarshalMap(t, parseMetadata),
	}, parseUnsignedResponse)

	// Test Combine
	// signaturesRaw := `[{"hex_bytes":"8c712c64bc65c4a88707fa93ecd090144dffb1bf133805a10a51d354c2f9f2b25a63cea6989f4c58372c41f31164036a6b25dce1d5c05e1d31c16c0590c176e801","signing_payload":{"address":"0xe3a5B4d7f79d64088C8d4ef153A7DDe2B2d47309","hex_bytes":"b682f3e39c512ff57471f482eab264551487320cbd3b34485f4779a89e5612d1","account_identifier":{"address":"0xe3a5B4d7f79d64088C8d4ef153A7DDe2B2d47309"},"signature_type":"ecdsa_recovery"},"public_key":{"hex_bytes":"03d3d3358e7f69cbe45bde38d7d6f24660c7eeeaee5c5590cfab985c8839b21fd5","curve_type":"secp256k1"},"signature_type":"ecdsa_recovery"}]` // nolint

	signaturesRaw := `[{"hex_bytes":"5f22dc4b318c51f636beb17e0483ca8f36d7a43d8acdff63eaed921bff5dc2c20f51930067cb001dbb1ba675e652cbf93375b4646b31d0818aa94917f3e6fda600","signing_payload":{"address":"0xbe862AD9AbFe6f22BCb087716c7D89a26051f74C","hex_bytes":"996836219400142b587d5ad87b1f70d25a2497fd6ac431509ed90c48df8b2a9f","account_identifier":{"address":"0xbe862AD9AbFe6f22BCb087716c7D89a26051f74C"},"signature_type":"ecdsa_recovery"},"public_key":{"hex_bytes":"036d9038945ff8f4669201ba1e806c9a46a5034a578e4d52c03152198538039294","curve_type":"secp256k1"},"signature_type":"ecdsa_recovery"}]` // nolint
	//tx: 0xaa26f7c0885128219e831a432ac58bc9ed79d26547ad65c27fd36e18c0ee232d
	//signaturesRaw := `[{"hex_bytes":"f17333e48f62ce7798119b91a7c0d523d2eecf74db0a7ef29ecb54a461f648c8685390d22e29d518bfea5d61cd7a14cb98a14fb2f8a8c7d335c1c3cec1b46f8a1c","signing_payload":{"address":"0x1B892F4bf95b25375D7E83A7D2E2641A4bdf3bfB","hex_bytes":"b682f3e39c512ff57471f482eab264551487320cbd3b34485f4779a89e5612d1","account_identifier":{"address":"0xe3a5B4d7f79d64088C8d4ef153A7DDe2B2d47309"},"signature_type":"ecdsa_recovery"},"public_key":{"hex_bytes":"03d3d3358e7f69cbe45bde38d7d6f24660c7eeeaee5c5590cfab985c8839b21fd5","curve_type":"secp256k1"},"signature_type":"ecdsa_recovery"}]` // nolint
	var signatures []*types.Signature
	assert.NoError(t, json.Unmarshal([]byte(signaturesRaw), &signatures))
	log(t, "signatures: ", signatures)
	// signedRaw := `{"type":"0x0","nonce":"0x0","gasPrice":"0x3b9aca00","maxPriorityFeePerGas":null,"maxFeePerGas":null,"gas":"0x5208","value":"0x9864aac3510d02","input":"0x","v":"0x2a","r":"0x8c712c64bc65c4a88707fa93ecd090144dffb1bf133805a10a51d354c2f9f2b2","s":"0x5a63cea6989f4c58372c41f31164036a6b25dce1d5c05e1d31c16c0590c176e8","to":"0x57b414a0332b5cab885a451c2a28a07d1e9b8a8d","hash":"0x424969b1a98757bcd748c60bad2a7de9745cfb26bfefb4550e780a098feada42"}` // nolint
	signedRaw := `{"type":"0x0","nonce":"0x0","gasPrice":"0x12a05f2000","maxPriorityFeePerGas":null,"maxFeePerGas":null,"gas":"0x5208","value":"0x9864aac3510d02","input":"0x","v":"0x8d3","r":"0x5f22dc4b318c51f636beb17e0483ca8f36d7a43d8acdff63eaed921bff5dc2c2","s":"0xf51930067cb001dbb1ba675e652cbf93375b4646b31d0818aa94917f3e6fda6","to":"0x57b414a0332b5cab885a451c2a28a07d1e9b8a8d","hash":"0x6e8d525fa1271b71f47e4f42bc2982ed7aecdfebfb56bc0d3d65cbf5521c9a3d"}` // nolint

	//tx: 0xaa26f7c0885128219e831a432ac58bc9ed79d26547ad65c27fd36e18c0ee232d
	//signedRaw := `{"type":"0x0","nonce": "0x3e2c4b","gasPrice":"0x12a05f2000","maxPriorityFeePerGas":null,"maxFeePerGas":null,"gas": "0x5208","value": "0x1","input": "0x","v": "0x1c","r": "0xf17333e48f62ce7798119b91a7c0d523d2eecf74db0a7ef29ecb54a461f648c8","s": "0x685390d22e29d518bfea5d61cd7a14cb98a14fb2f8a8c7d335c1c3cec1b46f8a","to": "0x2d74530c0c196de44d3906822053bf336f18a16e","hash": "0xaa26f7c0885128219e831a432ac58bc9ed79d26547ad65c27fd36e18c0ee232d"}`

	combineResponse, err := servicer.ConstructionCombine(ctx, &types.ConstructionCombineRequest{
		NetworkIdentifier:   networkIdentifier,
		UnsignedTransaction: unsignedRaw,
		Signatures:          signatures,
	})
	log(t, "combineResponse: ", combineResponse)
	log(t, "sample: ", &types.ConstructionCombineResponse{
		SignedTransaction: signedRaw,
	})
	assert.Nil(t, err)
	assert.Equal(t, &types.ConstructionCombineResponse{
		SignedTransaction: signedRaw,
	}, combineResponse)

	// Test Parse Signed
	parseSignedResponse, err := servicer.ConstructionParse(ctx, &types.ConstructionParseRequest{
		NetworkIdentifier: networkIdentifier,
		Signed:            true,
		Transaction:       signedRaw,
	})
	assert.Nil(t, err)
	log(t, "parseSignedResponse: ", parseSignedResponse)
	log(t, "sampe: ", &types.ConstructionParseResponse{
		Operations: parseOps,
		AccountIdentifierSigners: []*types.AccountIdentifier{
			{Address: "0xb22694a52EA2a9564001aF4AA61ecD9672E0D26b"},
		},
		Metadata: forceMarshalMap(t, parseMetadata),
	})
	assert.Equal(t, &types.ConstructionParseResponse{
		Operations: parseOps,
		AccountIdentifierSigners: []*types.AccountIdentifier{
			{Address: "0xb22694a52EA2a9564001aF4AA61ecD9672E0D26b"},
		},
		Metadata: forceMarshalMap(t, parseMetadata),
	}, parseSignedResponse)

	// Test Hash
	transactionIdentifier := &types.TransactionIdentifier{
		//	Hash: "0x424969b1a98757bcd748c60bad2a7de9745cfb26bfefb4550e780a098feada42",
		Hash: "0x6e8d525fa1271b71f47e4f42bc2982ed7aecdfebfb56bc0d3d65cbf5521c9a3d",
	}
	hashResponse, err := servicer.ConstructionHash(ctx, &types.ConstructionHashRequest{
		NetworkIdentifier: networkIdentifier,
		SignedTransaction: signedRaw,
	})
	assert.Nil(t, err)
	assert.Equal(t, &types.TransactionIdentifierResponse{
		TransactionIdentifier: transactionIdentifier,
	}, hashResponse)

	// Test Submit
	mockClient.On(
		"SendTransaction",
		ctx,
		mock.Anything, // can't test ethTx here because it contains "time"
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {

			tx := args.Get(1).(*ethTypes.Transaction)
			log(t, "tx: ", tx)
			data, err := rlp.EncodeToBytes(tx)
			log(t, "data: ", data)
			if err != nil {
				t.Error("Error EncodeToBytes")
			}
			t.Log("encode data:", hexutil.Encode(data))
			// return ec.c.CallContext(ctx, nil, "eth_sendRawTransaction", hexutil.Encode(data))
		},
	).Once()

	submitResponse, err := servicer.ConstructionSubmit(ctx, &types.ConstructionSubmitRequest{
		NetworkIdentifier: networkIdentifier,
		SignedTransaction: signedRaw,
	})

	log(t, "submitResponse: ", submitResponse)
	log(t, "sample: ", &types.TransactionIdentifierResponse{
		TransactionIdentifier: transactionIdentifier,
	})
	assert.Nil(t, err)
	assert.Equal(t, &types.TransactionIdentifierResponse{
		TransactionIdentifier: transactionIdentifier,
	}, submitResponse)

	mockClient.AssertExpectations(t)
}
