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

package wemix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"reflect"
	"sort"
	"testing"

	mocks "github.com/wemixarchive/rosetta-wemix/mocks/wemix"

	RosettaTypes "github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/sync/semaphore"
)

func TestStatus_NotReady(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"latest",
		false,
	).Return(
		nil,
	).Once()

	block, timestamp, syncStatus, peers, err := c.Status(ctx)
	assert.Nil(t, block)
	assert.Equal(t, int64(-1), timestamp)
	assert.Nil(t, syncStatus)
	assert.Nil(t, peers)
	assert.True(t, errors.Is(err, ethereum.NotFound))

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestStatus_NotSyncing(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"latest",
		false,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			header := args.Get(1).(**types.Header)
			file, err := ioutil.ReadFile("testdata/basic_header.json")
			assert.NoError(t, err)

			*header = new(types.Header)

			assert.NoError(t, (*header).UnmarshalJSON(file))
		},
	).Once()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_syncing",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			status := args.Get(1).(*json.RawMessage)

			*status = json.RawMessage("false")
		},
	).Once()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"admin_peers",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			info := args.Get(1).(*[]*p2p.PeerInfo)

			file, err := ioutil.ReadFile("testdata/peers.json")
			assert.NoError(t, err)

			assert.NoError(t, json.Unmarshal(file, info))
		},
	).Once()

	block, timestamp, syncStatus, peers, err := c.Status(ctx)
	tmp, _ := json.Marshal(peers)
	t.Log("peers :", string(tmp))
	assert.Equal(t, &RosettaTypes.BlockIdentifier{
		Hash:  "0xdabb8caa0c90b62db442f6725e57016d81f2a6d07394a110e839da12fe12965d",
		Index: 53151,
	}, block)
	assert.Equal(t, int64(1553151611000), timestamp)
	assert.Nil(t, syncStatus)
	assert.Equal(t, []*RosettaTypes.Peer{
		{
			PeerID: "30a7e74a861b54d40b5dc02a4aebdb5c5575dc5bba312873c763b4a8b5b7195d",
			Metadata: map[string]interface{}{
				"caps": []string{
					"wemix/63",
					"wemix/64",
					"wemix/65",
				},
				"enode": "enode://dc9c30053d98e55fe61bb2ef37d2f1be340bd295aa413749d0d6c76618050a358e2c737a904b2a472e6988322cffaebc8598fa59a9e19db99f1944ce1ebbbf89@52.68.113.68:8589", // nolint
				"enr":   "",
				"name":  "Gwemix/v0.9.7-stable-ff365d19/linux-amd64/go1.12",
				"protocols": map[string]interface{}{
					"wemix": map[string]interface{}{
						"difficulty": float64(19554721),
						"head":       "0x516ceed303d27630458c8ec8101132e72e1e01d2be0322c23618cd16fcfebb22",
						"version":    float64(65),
					},
				},
			},
		},
		{
			PeerID: "6b610d07d18d143a6c05de57fcccb06cb973efd29b1e6bce1b1629cbe342984a",
			Metadata: map[string]interface{}{
				"caps": []string{
					"wemix/63",
					"wemix/64",
					"wemix/65",
				},
				"enode": "enode://ce46b336757daf253ed7a89efa0f83af06cc37e01c14bc156e46557087a184567a67d20ef40b00f06c28953611b9b271c303039341f700dba5d3f5fd63f66d27@52.196.252.240:8589", // nolint
				"enr":   "",                                                                                                                                                             // nolint
				"name":  "Gwemix/v0.9.7-stable-ff365d19/linux-amd64/go1.12",
				"protocols": map[string]interface{}{
					"wemix": map[string]interface{}{
						"difficulty": float64(19554727),
						"head":       "0xe792cc88f659a4dae075de7e3ab7eca2ef8fc0eb83ecaf694552b87a3fc333fb",
						"version":    float64(65),
					},
				},
			},
		},
		{
			PeerID: "9b24b745e6d184ee97398a3af1a764132cf9ad9f0c5f9defd7a50e20eddcad6c",
			Metadata: map[string]interface{}{
				"caps": []string{
					"wemix/63",
					"wemix/64",
					"wemix/65",
				},
				"enode": "enode://a6d0067ef52e41e30e6417ba3fa15fdfcc820c47f0932eac6a659cdf9306443bbcd900e74710fbedd3c1cb50b4ef940fc944130345e7786816c1a8a14cda5aba@54.250.11.170:8589", // nolint
				"enr":   "",                                                                                                                                                            // nolint
				"name":  "Gwemix/v0.9.7-stable-ff365d19/linux-amd64/go1.12",
				"protocols": map[string]interface{}{
					"wemix": map[string]interface{}{
						"difficulty": float64(19554727),
						"head":       "0xe792cc88f659a4dae075de7e3ab7eca2ef8fc0eb83ecaf694552b87a3fc333fb",
						"version":    float64(65),
					},
				},
			},
		},
	}, peers)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestStatus_NotSyncing_SkipAdminCalls(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
		skipAdminCalls: true,
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"latest",
		false,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			header := args.Get(1).(**types.Header)
			file, err := ioutil.ReadFile("testdata/basic_header.json")
			assert.NoError(t, err)

			*header = new(types.Header)

			assert.NoError(t, (*header).UnmarshalJSON(file))
		},
	).Once()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_syncing",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			status := args.Get(1).(*json.RawMessage)

			*status = json.RawMessage("false")
		},
	).Once()

	adminPeersSkipped := true
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"admin_peers",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			adminPeersSkipped = false
		},
	).Maybe()

	block, timestamp, syncStatus, peers, err := c.Status(ctx)
	assert.True(t, adminPeersSkipped)
	assert.Equal(t, &RosettaTypes.BlockIdentifier{
		Hash:  "0xdabb8caa0c90b62db442f6725e57016d81f2a6d07394a110e839da12fe12965d",
		Index: 53151,
	}, block)
	assert.Equal(t, int64(1553151611000), timestamp)
	assert.Nil(t, syncStatus)
	assert.Equal(t, []*RosettaTypes.Peer{}, peers)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestStatus_Syncing(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"latest",
		false,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			header := args.Get(1).(**types.Header)
			file, err := ioutil.ReadFile("testdata/basic_header.json")
			assert.NoError(t, err)

			*header = new(types.Header)

			assert.NoError(t, (*header).UnmarshalJSON(file))
		},
	).Once()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_syncing",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			progress := args.Get(1).(*json.RawMessage)
			file, err := ioutil.ReadFile("testdata/syncing_info.json")
			assert.NoError(t, err)

			*progress = json.RawMessage(file)
		},
	).Once()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"admin_peers",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			info := args.Get(1).(*[]*p2p.PeerInfo)

			file, err := ioutil.ReadFile("testdata/peers.json")
			assert.NoError(t, err)

			assert.NoError(t, json.Unmarshal(file, info))
		},
	).Once()

	block, timestamp, syncStatus, peers, err := c.Status(ctx)
	assert.Equal(t, &RosettaTypes.BlockIdentifier{
		Hash:  "0xdabb8caa0c90b62db442f6725e57016d81f2a6d07394a110e839da12fe12965d",
		Index: 53151,
	}, block)
	assert.Equal(t, int64(1553151611000), timestamp)
	assert.Equal(t, &RosettaTypes.SyncStatus{
		CurrentIndex: RosettaTypes.Int64(8333),
		TargetIndex:  RosettaTypes.Int64(19554598),
	}, syncStatus)
	assert.Equal(t, []*RosettaTypes.Peer{
		{
			PeerID: "30a7e74a861b54d40b5dc02a4aebdb5c5575dc5bba312873c763b4a8b5b7195d",
			Metadata: map[string]interface{}{
				"caps": []string{
					"wemix/63",
					"wemix/64",
					"wemix/65",
				},
				"enode": "enode://dc9c30053d98e55fe61bb2ef37d2f1be340bd295aa413749d0d6c76618050a358e2c737a904b2a472e6988322cffaebc8598fa59a9e19db99f1944ce1ebbbf89@52.68.113.68:8589", // nolint
				"enr":   "",
				"name":  "Gwemix/v0.9.7-stable-ff365d19/linux-amd64/go1.12",
				"protocols": map[string]interface{}{
					"wemix": map[string]interface{}{
						"difficulty": float64(19554721),
						"head":       "0x516ceed303d27630458c8ec8101132e72e1e01d2be0322c23618cd16fcfebb22",
						"version":    float64(65),
					},
				},
			},
		},
		{
			PeerID: "6b610d07d18d143a6c05de57fcccb06cb973efd29b1e6bce1b1629cbe342984a",
			Metadata: map[string]interface{}{
				"caps": []string{
					"wemix/63",
					"wemix/64",
					"wemix/65",
				},
				"enode": "enode://ce46b336757daf253ed7a89efa0f83af06cc37e01c14bc156e46557087a184567a67d20ef40b00f06c28953611b9b271c303039341f700dba5d3f5fd63f66d27@52.196.252.240:8589", // nolint
				"enr":   "",                                                                                                                                                             // nolint
				"name":  "Gwemix/v0.9.7-stable-ff365d19/linux-amd64/go1.12",
				"protocols": map[string]interface{}{
					"wemix": map[string]interface{}{
						"difficulty": float64(19554727),
						"head":       "0xe792cc88f659a4dae075de7e3ab7eca2ef8fc0eb83ecaf694552b87a3fc333fb",
						"version":    float64(65),
					},
				},
			},
		},
		{
			PeerID: "9b24b745e6d184ee97398a3af1a764132cf9ad9f0c5f9defd7a50e20eddcad6c",
			Metadata: map[string]interface{}{
				"caps": []string{
					"wemix/63",
					"wemix/64",
					"wemix/65",
				},
				"enode": "enode://a6d0067ef52e41e30e6417ba3fa15fdfcc820c47f0932eac6a659cdf9306443bbcd900e74710fbedd3c1cb50b4ef940fc944130345e7786816c1a8a14cda5aba@54.250.11.170:8589", // nolint
				"enr":   "",                                                                                                                                                            // nolint
				"name":  "Gwemix/v0.9.7-stable-ff365d19/linux-amd64/go1.12",
				"protocols": map[string]interface{}{
					"wemix": map[string]interface{}{
						"difficulty": float64(19554727),
						"head":       "0xe792cc88f659a4dae075de7e3ab7eca2ef8fc0eb83ecaf694552b87a3fc333fb",
						"version":    float64(65),
					},
				},
			},
		},
	}, peers)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestStatus_Syncing_SkipAdminCalls(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
		skipAdminCalls: true,
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"latest",
		false,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			header := args.Get(1).(**types.Header)
			file, err := ioutil.ReadFile("testdata/basic_header.json")
			assert.NoError(t, err)

			*header = new(types.Header)

			assert.NoError(t, (*header).UnmarshalJSON(file))
		},
	).Once()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_syncing",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			progress := args.Get(1).(*json.RawMessage)
			file, err := ioutil.ReadFile("testdata/syncing_info.json")
			assert.NoError(t, err)

			*progress = json.RawMessage(file)
		},
	).Once()

	adminPeersSkipped := true
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"admin_peers",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			adminPeersSkipped = false
		},
	).Maybe()

	block, timestamp, syncStatus, peers, err := c.Status(ctx)
	assert.True(t, adminPeersSkipped)
	assert.Equal(t, &RosettaTypes.BlockIdentifier{
		Hash:  "0xdabb8caa0c90b62db442f6725e57016d81f2a6d07394a110e839da12fe12965d",
		Index: 53151,
	}, block)
	assert.Equal(t, int64(1553151611000), timestamp)
	assert.Equal(t, &RosettaTypes.SyncStatus{
		CurrentIndex: RosettaTypes.Int64(8333),
		TargetIndex:  RosettaTypes.Int64(19554598),
	}, syncStatus)
	assert.Equal(t, []*RosettaTypes.Peer{}, peers)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestBalance(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	result, err := ioutil.ReadFile(
		"testdata/account_balance_0x098cE27428a8fe633f1177f8253Ea789894d8aDf.json",
	)
	assert.NoError(t, err)
	mockGraphQL.On(
		"Query",
		ctx,
		`{
			block(){
				hash
				number
				account(address:"0x2f93B2f047E05cdf602820Ac4B3178efc2b43D55"){
					balance
					transactionCount
					code
				}
			}
		}`,
	).Return(
		string(result),
		nil,
	).Once()

	resp, err := c.Balance(
		ctx,
		&RosettaTypes.AccountIdentifier{
			Address: "0x2f93B2f047E05cdf602820Ac4B3178efc2b43D55",
		},
		nil,
	)

	assert.Equal(t, &RosettaTypes.AccountBalanceResponse{
		BlockIdentifier: &RosettaTypes.BlockIdentifier{
			Hash:  "0xb72ccf3ec2617015c9c3751c66a4918a7fd9a0d1667ac7cadb1601a6f442889d",
			Index: 19388485,
		},
		Balances: []*RosettaTypes.Amount{
			{
				Value:    "1390630720000000000",
				Currency: Currency,
			},
		},
		Metadata: map[string]interface{}{
			"code":  "0x",
			"nonce": int64(0),
		},
	}, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestBalance_Historical_Hash(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	result, err := ioutil.ReadFile(
		"testdata/account_balance_0x098cE27428a8fe633f1177f8253Ea789894d8aDf.json",
	)
	assert.NoError(t, err)
	mockGraphQL.On(
		"Query",
		ctx,
		`{
			block(hash: "0xb72ccf3ec2617015c9c3751c66a4918a7fd9a0d1667ac7cadb1601a6f442889d"){
				hash
				number
				account(address:"0x2f93B2f047E05cdf602820Ac4B3178efc2b43D55"){
					balance
					transactionCount
					code
				}
			}
		}`,
	).Return(
		string(result),
		nil,
	).Once()

	resp, err := c.Balance(
		ctx,
		&RosettaTypes.AccountIdentifier{
			Address: "0x2f93B2f047E05cdf602820Ac4B3178efc2b43D55",
		},
		&RosettaTypes.PartialBlockIdentifier{
			Hash: RosettaTypes.String(
				"0xb72ccf3ec2617015c9c3751c66a4918a7fd9a0d1667ac7cadb1601a6f442889d",
			),
			Index: RosettaTypes.Int64(19388485),
		},
	)
	assert.Equal(t, &RosettaTypes.AccountBalanceResponse{
		BlockIdentifier: &RosettaTypes.BlockIdentifier{
			Hash:  "0xb72ccf3ec2617015c9c3751c66a4918a7fd9a0d1667ac7cadb1601a6f442889d",
			Index: 19388485,
		},
		Balances: []*RosettaTypes.Amount{
			{
				Value:    "1390630720000000000",
				Currency: Currency,
			},
		},
		Metadata: map[string]interface{}{
			"code":  "0x",
			"nonce": int64(0),
		},
	}, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestBalance_Historical_Index(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	result, err := ioutil.ReadFile(
		"testdata/account_balance_0x098cE27428a8fe633f1177f8253Ea789894d8aDf.json",
	)
	assert.NoError(t, err)
	mockGraphQL.On(
		"Query",
		ctx,
		`{
			block(number: 19388485){
				hash
				number
				account(address:"0x2f93B2f047E05cdf602820Ac4B3178efc2b43D55"){
					balance
					transactionCount
					code
				}
			}
		}`,
	).Return(
		string(result),
		nil,
	).Once()

	resp, err := c.Balance(
		ctx,
		&RosettaTypes.AccountIdentifier{
			Address: "0x2f93B2f047E05cdf602820Ac4B3178efc2b43D55",
		},
		&RosettaTypes.PartialBlockIdentifier{
			Index: RosettaTypes.Int64(19388485),
		},
	)
	assert.Equal(t, &RosettaTypes.AccountBalanceResponse{
		BlockIdentifier: &RosettaTypes.BlockIdentifier{
			Hash:  "0xb72ccf3ec2617015c9c3751c66a4918a7fd9a0d1667ac7cadb1601a6f442889d",
			Index: 19388485,
		},
		Balances: []*RosettaTypes.Amount{
			{
				Value:    "1390630720000000000",
				Currency: Currency,
			},
		},
		Metadata: map[string]interface{}{
			"code":  "0x",
			"nonce": int64(0),
		},
	}, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestBalance_InvalidAddress(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	result, err := ioutil.ReadFile("testdata/account_balance_invalid.json")
	assert.NoError(t, err)
	mockGraphQL.On(
		"Query",
		ctx,
		`{
			block(){
				hash
				number
				account(address:"0x4cfc400fed52f9681b42454c2db4b18ab98f8de"){
					balance
					transactionCount
					code
				}
			}
		}`,
	).Return(
		string(result),
		nil,
	).Once()

	resp, err := c.Balance(
		ctx,
		&RosettaTypes.AccountIdentifier{
			Address: "0x4cfc400fed52f9681b42454c2db4b18ab98f8de",
		},
		nil,
	)
	assert.Nil(t, resp)
	assert.Error(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestBalance_InvalidHash(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	result, err := ioutil.ReadFile("testdata/account_balance_invalid_block.json")
	assert.NoError(t, err)
	mockGraphQL.On(
		"Query",
		ctx,
		`{
			block(hash: "0x7d2a2713026a0e66f131878de2bb2df2fff6c24562c1df61ec0265e5fedf2626"){
				hash
				number
				account(address:"0x2f93B2f047E05cdf602820Ac4B3178efc2b43D55"){
					balance
					transactionCount
					code
				}
			}
		}`,
	).Return(
		string(result),
		nil,
	).Once()

	resp, err := c.Balance(
		ctx,
		&RosettaTypes.AccountIdentifier{
			Address: "0x2f93B2f047E05cdf602820Ac4B3178efc2b43D55",
		},
		&RosettaTypes.PartialBlockIdentifier{
			Hash: RosettaTypes.String(
				"0x7d2a2713026a0e66f131878de2bb2df2fff6c24562c1df61ec0265e5fedf2626",
			),
		},
	)
	assert.Nil(t, resp)
	assert.Error(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_GetBlockByNumber(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"0x2af0",
		false,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*map[string]interface{})

			file, err := ioutil.ReadFile("testdata/block_10992.json")
			assert.NoError(t, err)

			err = json.Unmarshal(file, r)
			assert.NoError(t, err)
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_10992.json")
	assert.NoError(t, err)
	var correct map[string]interface{}
	assert.NoError(t, json.Unmarshal(correctRaw, &correct))

	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "eth_getBlockByNumber",
			Parameters: map[string]interface{}{
				"index":                    RosettaTypes.Int64(10992),
				"show_transaction_details": false,
			},
		},
	)
	tmp, err := json.Marshal(resp)
	t.Log("resp :", string(tmp))

	assert.Equal(t, &RosettaTypes.CallResponse{
		Result:     correct,
		Idempotent: false,
	}, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_GetBlockByNumber_InvalidArgs(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "eth_getBlockByNumber",
			Parameters: map[string]interface{}{
				"index":                    "a string",
				"show_transaction_details": false,
			},
		},
	)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, ErrCallParametersInvalid))

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_GetTransactionReceipt(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getTransactionReceipt",
		common.HexToHash("0xe8f6e3d60dd6ec3b28aa2f43a915d28924fc458c914d2ac0d52c7ec9e153268a"),
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(**types.Receipt)

			file, err := ioutil.ReadFile(
				"testdata/call_0xe8f6e3d60dd6ec3b28aa2f43a915d28924fc458c914d2ac0d52c7ec9e153268a.json",
			)
			assert.NoError(t, err)
			t.Log("file :", string(file))
			*r = new(types.Receipt)

			assert.NoError(t, (*r).UnmarshalJSON(file))
			tmp, _ := json.Marshal(r)
			t.Log("Receipt :", string(tmp))
		},
	).Once()
	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "eth_getTransactionReceipt",
			Parameters: map[string]interface{}{
				"tx_hash": "0xe8f6e3d60dd6ec3b28aa2f43a915d28924fc458c914d2ac0d52c7ec9e153268a",
			},
		},
	)
	tmp, _ := json.Marshal(resp)
	t.Log("resp :", string(tmp))

	assert.Equal(t, &RosettaTypes.CallResponse{
		Result: map[string]interface{}{
			"blockHash":         "0xafba9f1d5b50f18391e440fe4d4578e6476e0889a365317182aa9c83dd7eabce",
			"blockNumber":       "0x128e142",
			"contractAddress":   "0x0000000000000000000000000000000000000000",
			"cumulativeGasUsed": "0x5208",
			"gasUsed":           "0x5208",
			"logs":              []interface{}{},
			"logsBloom":         "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", // nolint
			"root":              "0x",
			"status":            "0x1",
			"transactionHash":   "0xe8f6e3d60dd6ec3b28aa2f43a915d28924fc458c914d2ac0d52c7ec9e153268a",
			"transactionIndex":  "0x0",
		},
		Idempotent: false,
	}, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_GetTransactionReceipt_InvalidArgs(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "eth_getTransactionReceipt",
		},
	)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, ErrCallParametersInvalid))

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_Call(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_call",
		map[string]string{
			"to":   "0xB5E5D0F8C0cbA267CD3D7035d6AdC8eBA7Df7Cdd",
			"data": "0x70a08231000000000000000000000000b5e5d0f8c0cba267cd3d7035d6adc8eba7df7cdd",
		},
		toBlockNumArg(big.NewInt(11408349)),
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*string)

			var expected map[string]interface{}
			file, err := ioutil.ReadFile("testdata/call_balance_11408349.json")
			assert.NoError(t, err)

			err = json.Unmarshal(file, &expected)
			assert.NoError(t, err)

			*r = expected["data"].(string)
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/call_balance_11408349.json")
	assert.NoError(t, err)
	var correct map[string]interface{}
	assert.NoError(t, json.Unmarshal(correctRaw, &correct))

	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "eth_call",
			Parameters: map[string]interface{}{
				"index": 11408349,
				"to":    "0xB5E5D0F8C0cbA267CD3D7035d6AdC8eBA7Df7Cdd",
				"data":  "0x70a08231000000000000000000000000b5e5d0f8c0cba267cd3d7035d6adc8eba7df7cdd",
			},
		},
	)
	assert.Equal(t, &RosettaTypes.CallResponse{
		Result:     correct,
		Idempotent: false,
	}, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_Call_InvalidArgs(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "eth_call",
			Parameters: map[string]interface{}{
				"index": 11408349,
				"Hash":  "0x73fc065bc04f16c98247f8ec1e990f581ec58723bcd8059de85f93ab18706448",
				"to":    "not valid  ",
				"data":  "0x70a08231000000000000000000000000b5e5d0f8c0cba267cd3d7035d6adc8eba7df7cdd",
			},
		},
	)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, ErrCallParametersInvalid))

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_EstimateGas(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_estimateGas",
		map[string]string{
			"from": "0xE550f300E477C60CE7e7172d12e5a27e9379D2e3",
			"to":   "0xaD6D458402F60fD3Bd25163575031ACDce07538D",
			"data": "0xa9059cbb000000000000000000000000ae7e48ee0f758cd706b76cf7e2175d982800879a" +
				"00000000000000000000000000000000000000000000000000521c5f98b8ea00",
		},
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*string)

			var expected map[string]interface{}
			file, err := ioutil.ReadFile("testdata/estimate_gas_0xaD6D458402F60fD3Bd25163575031ACDce07538D.json")
			assert.NoError(t, err)

			err = json.Unmarshal(file, &expected)
			assert.NoError(t, err)

			*r = expected["data"].(string)
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/estimate_gas_0xaD6D458402F60fD3Bd25163575031ACDce07538D.json")
	assert.NoError(t, err)
	var correct map[string]interface{}
	assert.NoError(t, json.Unmarshal(correctRaw, &correct))

	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "eth_estimateGas",
			Parameters: map[string]interface{}{
				"from": "0xE550f300E477C60CE7e7172d12e5a27e9379D2e3",
				"to":   "0xaD6D458402F60fD3Bd25163575031ACDce07538D",
				"data": "0xa9059cbb000000000000000000000000ae7e48ee0f758cd706b76cf7e2175d982800879a" +
					"00000000000000000000000000000000000000000000000000521c5f98b8ea00",
			},
		},
	)
	assert.Equal(t, &RosettaTypes.CallResponse{
		Result:     correct,
		Idempotent: false,
	}, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_EstimateGas_InvalidArgs(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "eth_estimateGas",
			Parameters: map[string]interface{}{
				"From": "0xE550f300E477C60CE7e7172d12e5a27e9379D2e3",
				"to":   "0xaD6D458402F60fD3Bd25163575031ACDce07538D",
			},
		},
	)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, ErrCallParametersInvalid))

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_InvalidMethod(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "blah",
		},
	)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, ErrCallMethodInvalid))

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func testTraceConfig() (*tracers.TraceConfig, error) {
	loadedFile, err := ioutil.ReadFile("call_tracer.js")
	if err != nil {
		return nil, fmt.Errorf("%w: could not load tracer file", err)
	}

	loadedTracer := string(loadedFile)
	return &tracers.TraceConfig{
		Timeout: &tracerTimeout,
		Tracer:  &loadedTracer,
	}, nil
}

func TestBlock_Current(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.WemixTestnetChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"latest",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_10992.json")
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceBlockByHash",
		common.HexToHash("0xf6240d887224149baf5c5bfa3836ff4d64faa0ef65c2c8cbb0b6a6106eb0c8bd"),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/block_trace_0xf6240d887224149baf5c5bfa3836ff4d64faa0ef65c2c8cbb0b6a6106eb0c8bd.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_10992.json")
	assert.NoError(t, err)
	var correct *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correct))

	resp, err := c.Block(
		ctx,
		nil,
	)
	tmp, err := json.Marshal(resp)
	t.Log("resp :", string(tmp))
	assert.Equal(t, correct.Block, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestBlock_Hash(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.WemixTestnetChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByHash",
		"0xf6240d887224149baf5c5bfa3836ff4d64faa0ef65c2c8cbb0b6a6106eb0c8bd",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_10992.json")
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceBlockByHash",
		common.HexToHash("0xf6240d887224149baf5c5bfa3836ff4d64faa0ef65c2c8cbb0b6a6106eb0c8bd"),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/block_trace_0xf6240d887224149baf5c5bfa3836ff4d64faa0ef65c2c8cbb0b6a6106eb0c8bd.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_10992.json")
	assert.NoError(t, err)
	var correct *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correct))

	resp, err := c.Block(
		ctx,
		&RosettaTypes.PartialBlockIdentifier{
			Hash: RosettaTypes.String(
				"0xf6240d887224149baf5c5bfa3836ff4d64faa0ef65c2c8cbb0b6a6106eb0c8bd",
			),
		},
	)
	assert.Equal(t, correct.Block, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestBlock_Index(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.WemixTestnetChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"0x2af0",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_10992.json")
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceBlockByHash",
		common.HexToHash("0xf6240d887224149baf5c5bfa3836ff4d64faa0ef65c2c8cbb0b6a6106eb0c8bd"),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/block_trace_0xf6240d887224149baf5c5bfa3836ff4d64faa0ef65c2c8cbb0b6a6106eb0c8bd.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_10992.json")
	assert.NoError(t, err)
	var correct *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correct))

	resp, err := c.Block(
		ctx,
		&RosettaTypes.PartialBlockIdentifier{
			Index: RosettaTypes.Int64(10992),
		},
	)
	assert.Equal(t, correct.Block, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestBlock_FirstBlock(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.RopstenChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"0x0",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_0.json")
			assert.NoError(t, err)

			*r = file
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_0.json")
	assert.NoError(t, err)
	var correct *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correct))

	resp, err := c.Block(
		ctx,
		&RosettaTypes.PartialBlockIdentifier{
			Index: RosettaTypes.Int64(0),
		},
	)
	tmp, err := json.Marshal(resp)
	tmp2, err := json.Marshal(correct.Block)
	t.Log("resp :", string(tmp))
	t.Log("correct.Block :", string(tmp2))
	assert.Equal(t, correct.Block, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func jsonifyTransaction(b *RosettaTypes.Transaction) (*RosettaTypes.Transaction, error) {
	bytes, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	var tx RosettaTypes.Transaction
	if err := json.Unmarshal(bytes, &tx); err != nil {
		return nil, err
	}

	return &tx, nil
}

func jsonifyBlock(b *RosettaTypes.Block) (*RosettaTypes.Block, error) {
	bytes, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	var bo RosettaTypes.Block
	if err := json.Unmarshal(bytes, &bo); err != nil {
		return nil, err
	}

	return &bo, nil
}

func TestTransaction_Hash(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}
	txHash := "0x9cc8e6a09ae9cbdb7da77515110a8e343a945df4269c53842dd26969d32c6cc4"
	blockHash := "0xc10a51a3898a85c7165a9d883acc9a68f139934d0cb91dfad4c7d3a7c1a1960d"

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.RopstenChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getTransactionByHash",
		txHash,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/transaction_0x9cc8e6a09ae9cbdb7da77515110a8e343a945df4269c53842dd26969d32c6cc4.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByHash",
		blockHash,
		false,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(**types.Header)

			file, err := ioutil.ReadFile(
				"testdata/block_0xc10a51a3898a85c7165a9d883acc9a68f139934d0cb91dfad4c7d3a7c1a1960d.json",
			) // nolint
			assert.NoError(t, err)

			*r = new(types.Header)

			assert.NoError(t, (*r).UnmarshalJSON(file))
		},
	).Once()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getTransactionReceipt",
		common.HexToHash(txHash),
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(**types.Receipt)

			file, err := ioutil.ReadFile(
				"testdata/tx_receipt_0x9cc8e6a09ae9cbdb7da77515110a8e343a945df4269c53842dd26969d32c6cc4.json",
			) // nolint
			assert.NoError(t, err)

			*r = new(types.Receipt)

			assert.NoError(t, (*r).UnmarshalJSON(file))
		},
	).Once()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceTransaction",
		common.HexToHash(txHash),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/transaction_trace_0x9cc8e6a09ae9cbdb7da77515110a8e343a945df4269c53842dd26969d32c6cc4.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()

	correctRaw, err := ioutil.ReadFile(
		"testdata/transaction_response_0x9cc8e6a09ae9cbdb7da77515110a8e343a945df4269c53842dd26969d32c6cc4.json",
	) // nolint
	assert.NoError(t, err)
	var correct *RosettaTypes.BlockTransactionResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correct))

	resp, err := c.Transaction(
		ctx,
		&RosettaTypes.BlockIdentifier{
			Hash: "0xc10a51a3898a85c7165a9d883acc9a68f139934d0cb91dfad4c7d3a7c1a1960d",
		},
		&RosettaTypes.TransactionIdentifier{
			Hash: "0x9cc8e6a09ae9cbdb7da77515110a8e343a945df4269c53842dd26969d32c6cc4",
		},
	)
	assert.NoError(t, err)

	jsonResp, err := jsonifyTransaction(resp)
	assert.NoError(t, err)
	assert.Equal(t, correct.Transaction, jsonResp)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

// Block with transaction
// func TestBlock_10994(t *testing.T) {
func TestBlock_14497230(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.WemixTestnetChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"0xdd35ce",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_14497230.json")
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceBlockByHash",
		common.HexToHash("0xacccbfcbe791d0e15c6797ccc72d1f6bb0948d3bc6f738f38dd642c323513b0d"),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/block_trace_0xacccbfcbe791d0e15c6797ccc72d1f6bb0948d3bc6f738f38dd642c323513b0d.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"BatchCallContext",
		ctx,
		mock.Anything,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).([]rpc.BatchElem)

			assert.Len(t, r, 1)
			assert.Equal(
				t,
				"0x47d4a3a76e13d96aa898e313ccb941966373dd9f9c668535e5a8f49c137af5b2",
				r[0].Args[0],
			)

			file, err := ioutil.ReadFile(
				"testdata/tx_receipt_0x47d4a3a76e13d96aa898e313ccb941966373dd9f9c668535e5a8f49c137af5b2.json",
			) // nolint
			assert.NoError(t, err)

			receipt := new(types.Receipt)
			assert.NoError(t, receipt.UnmarshalJSON(file))
			*(r[0].Result.(**types.Receipt)) = receipt
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_14497230.json")
	assert.NoError(t, err)
	var correctResp *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correctResp))

	resp, err := c.Block(
		ctx,
		&RosettaTypes.PartialBlockIdentifier{
			Index: RosettaTypes.Int64(14497230),
		},
	)
	assert.NoError(t, err)

	// Ensure types match
	jsonResp, err := jsonifyBlock(resp)
	tmp, _ := json.Marshal(jsonResp)
	tmp2, _ := json.Marshal(correctResp.Block)
	t.Log("resp :", string(tmp))
	t.Log("tmp2 :", string(tmp2))
	assert.NoError(t, err)
	assert.Equal(t, correctResp.Block, jsonResp)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

// Block with partial success transaction
func TestBlock_239782(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.RopstenChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"0x3a8a6",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_239782.json")
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceBlockByHash",
		common.HexToHash("0xc4487850a40d85b79cf5e5b69db38284fbd39efcf902ca8a6d9f2ba89c538ea3"),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/block_trace_0xc4487850a40d85b79cf5e5b69db38284fbd39efcf902ca8a6d9f2ba89c538ea3.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"BatchCallContext",
		ctx,
		mock.Anything,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).([]rpc.BatchElem)

			assert.Len(t, r, 1)
			assert.Equal(
				t,
				"0x05613760334d347e771fad61b1815c8c817b8dd5f0fcbba57c3f2df67dec33d6",
				r[0].Args[0],
			)

			file, err := ioutil.ReadFile(
				"testdata/tx_receipt_0x05613760334d347e771fad61b1815c8c817b8dd5f0fcbba57c3f2df67dec33d6.json",
			) // nolint
			assert.NoError(t, err)

			receipt := new(types.Receipt)
			assert.NoError(t, receipt.UnmarshalJSON(file))
			*(r[0].Result.(**types.Receipt)) = receipt
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_239782.json")
	assert.NoError(t, err)
	var correctResp *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correctResp))

	resp, err := c.Block(
		ctx,
		&RosettaTypes.PartialBlockIdentifier{
			Index: RosettaTypes.Int64(239782),
		},
	)
	assert.NoError(t, err)

	// Ensure types match
	jsonResp, err := jsonifyBlock(resp)

	tmp, _ := json.Marshal(jsonResp)
	t.Log("jsonResp :", string(tmp))
	tmp2, _ := json.Marshal(correctResp.Block)
	t.Log("correctResp.Block :", string(tmp2))

	assert.NoError(t, err)
	assert.Equal(t, correctResp.Block, jsonResp)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

// Block with EIP-1559 base fee & txs. This block taken from mainnet:
//
//	https://etherscan.io/block/0x68985b6b06bb5c6012393145729babb983fc16c50ec5207972ddda02de02f7e2
//
// This block has 7 transactions, all EIP-1559 type except the last.
func TestBlock_13998626(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.RopstenChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"0xd59a22",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_13998626.json")
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceBlockByHash",
		common.HexToHash("0x68985b6b06bb5c6012393145729babb983fc16c50ec5207972ddda02de02f7e2"),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/block_trace_0x68985b6b06bb5c6012393145729babb983fc16c50ec5207972ddda02de02f7e2.json") // nolint
			assert.NoError(t, err)
			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"BatchCallContext",
		ctx,
		mock.Anything,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).([]rpc.BatchElem)
			assert.Len(t, r, 7)
			for i, txHash := range []string{
				"0xf121c8c07ed51b6ac2d11fe3f0892bff2221ec9168280d12581ea8ff45e71421",
				"0xef0748860f1c1ba28a5ae3ae9d2d1133940f7c8090fc862acf48de42b00ae2b5",
				"0xb240b922161bb0aeaa5ebe67e6cf77311092bd945b9582b8deba61e2ebdde74f",
				"0xfac8149f95c20f62264991fe15dc74ca77c92ad6e4329496548277fb4d520509",
				"0x0a4cd36d72c2ed4767c1d228a7aa0638c3e46397f48b6b09f35ed455c851bb04",
				"0x9ee03d5922b2a901e3fc05d8a6351165b9f211162363c790c98746ef229e395c",
				"0x0d4a4f924858a5b19f6b931a914701d4258e73fa738da3d38eb3be1d1e862a7a",
			} {
				assert.Equal(
					t,
					txHash,
					r[i].Args[0],
				)

				file, err := ioutil.ReadFile(
					"testdata/tx_receipt_" + txHash + ".json",
				) // nolint
				assert.NoError(t, err)

				receipt := new(types.Receipt)
				assert.NoError(t, receipt.UnmarshalJSON(file))
				*(r[i].Result.(**types.Receipt)) = receipt
			}
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_13998626.json")
	assert.NoError(t, err)
	var correctResp *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correctResp))

	resp, err := c.Block(
		ctx,
		&RosettaTypes.PartialBlockIdentifier{
			Index: RosettaTypes.Int64(13998626),
		},
	)
	assert.NoError(t, err)

	// Ensure types match
	jsonResp, err := jsonifyBlock(resp)

	assert.NoError(t, err)
	assert.Equal(t, correctResp.Block, jsonResp)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestPendingNonceAt(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getTransactionCount",
		common.HexToAddress("0xfFC614eE978630D7fB0C06758DeB580c152154d3"),
		"pending",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*hexutil.Uint64)

			*r = hexutil.Uint64(10)
		},
	).Once()
	resp, err := c.PendingNonceAt(
		ctx,
		common.HexToAddress("0xfFC614eE978630D7fB0C06758DeB580c152154d3"),
	)
	assert.Equal(t, uint64(10), resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestSuggestGasPrice(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_gasPrice",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*hexutil.Big)

			*r = *(*hexutil.Big)(big.NewInt(80000000000))
		},
	).Once()
	resp, err := c.SuggestGasPrice(
		ctx,
	)
	assert.Equal(t, big.NewInt(80000000000), resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestSendTransaction(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_sendRawTransaction",
		"0xf86a80843b9aca00825208941ff502f9fe838cd772874cb67d0d96b93fd1d6d78725d4b6199a415d8029a01d110bf9fd468f7d00b3ce530832e99818835f45e9b08c66f8d9722264bb36c7a02711f47ec99f9ac585840daef41b7118b52ec72f02fcb30d874d36b10b668b59", // nolint
	).Return(
		nil,
	).Once()

	rawTx, err := ioutil.ReadFile("testdata/submitted_tx.json")
	assert.NoError(t, err)

	tx := new(types.Transaction)
	assert.NoError(t, tx.UnmarshalJSON(rawTx))

	assert.NoError(t, c.SendTransaction(
		ctx,
		tx,
	))

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestGetMempool(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}
	ctx := context.Background()
	expectedMempool := &RosettaTypes.MempoolResponse{
		TransactionIdentifiers: []*RosettaTypes.TransactionIdentifier{
			{Hash: "0x994024ef9f05d1cb25d01572642c1f550c78d214a52c306bb100d22c025b59d4"},
			{Hash: "0x0859cb844087a280fa031ce2a0879f4f9832f431d6de67f57d0ca32e90dd9e21"},
			{Hash: "0xa6de13e0a4465c9b55726d0826d020ed179fa1bda882a00e90aa467266af1815"},
			{Hash: "0x94e28f01121fd56dead7b89c46c8d6ba32bb79dad5f7e29c1d523224479f10f4"},
			{Hash: "0xf98b8a6589fa27f66545f31d443140c16eaa3da3896e66c4f05a03defa12cda4"},
			{Hash: "0x4c652b9e779277f06e17b6a4c4db6658ac012f7c93f5eeab80e7802cce3ff556"},
			{Hash: "0xa81f636c4b47831efd324dce5c48fe3a3d7a4f0020fae063887e8c2765ff1088"},
			{Hash: "0x3bc19098a49974002ba933ac8f9c753ce5af538ab93fc68586849e8b8b0dce80"},
			{Hash: "0x3401ec802cbe4b02d5be18717b53bb8177bd746f95a54da4674d7c27620facda"},
			{Hash: "0x3d1c58eced8f15f3224e9b2579420d078dbd2adba8e825cee91fd306b0943f89"},
			{Hash: "0xda591f0b15423aedb52f6b0e778b1fbc6757547d69277e2ba1aa7093583d1efb"},
			{Hash: "0xb39652e66c57a6693e44b92b5c471952c26cabf9442a9a76c372f8690658cb4c"},
			{Hash: "0x1e8700bf7215b2da0cfadcf34717f387577254e0871d828e5453a0593ce8060f"},
			{Hash: "0x83811383fb7840a03c25a7cbae7e9af138b17853563eb9e212727be2d0b9667f"},
			{Hash: "0x1e53751e1312cae3324a6b36c67dc95bfec993d7b4939c0de8c0dc761a0afd31"},
			{Hash: "0xc1052f9378db5a779c42ae2de9a0b94c8a6357c815446d6ba55485dcc1b187ef"},
		},
	}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	mockJSONRPC.On(
		"CallContext", ctx, mock.Anything, "txpool_content",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r, ok := args.Get(1).(*txPoolContentResponse)
			assert.True(t, ok)

			file, err := ioutil.ReadFile("testdata/txpool_content.json")
			assert.NoError(t, err)

			err = json.Unmarshal(file, r)
			assert.NoError(t, err)
		},
	).Once()

	actualMempool, err := c.GetMempool(ctx)
	assert.NoError(t, err)

	assert.Len(t, actualMempool.TransactionIdentifiers, len(expectedMempool.TransactionIdentifiers))

	// Sort both slices to compare later
	sort.Slice(expectedMempool.TransactionIdentifiers, func(i, j int) bool {
		return expectedMempool.TransactionIdentifiers[i].Hash < expectedMempool.TransactionIdentifiers[j].Hash
	})

	sort.Slice(actualMempool.TransactionIdentifiers, func(i, j int) bool {
		return actualMempool.TransactionIdentifiers[i].Hash < actualMempool.TransactionIdentifiers[j].Hash
	})

	assert.True(t, reflect.DeepEqual(actualMempool, expectedMempool))

	mockJSONRPC.AssertExpectations(t)
}
