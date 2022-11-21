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
	"testing"

	"github.com/wemixarchive/rosetta-wemix/configuration"
	mocks "github.com/wemixarchive/rosetta-wemix/mocks/services"
	"github.com/wemixarchive/rosetta-wemix/wemix"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"
)

var (
	middlewareVersion     = "0.0.4"
	defaultNetworkOptions = &types.NetworkOptionsResponse{
		Version: &types.Version{
			RosettaVersion:    types.RosettaAPIVersion,
			NodeVersion:       "m0.9.7",
			MiddlewareVersion: &middlewareVersion,
		},
		Allow: &types.Allow{
			OperationStatuses:       wemix.OperationStatuses,
			OperationTypes:          wemix.OperationTypes,
			Errors:                  Errors,
			HistoricalBalanceLookup: wemix.HistoricalBalanceSupported,
			CallMethods:             wemix.CallMethods,
		},
	}

	networkIdentifier = &types.NetworkIdentifier{
		Network:    wemix.MainnetNetwork,
		Blockchain: wemix.Blockchain,
	}
)

func TestNetworkEndpoints_Offline(t *testing.T) {
	cfg := &configuration.Configuration{
		Mode:    configuration.Offline,
		Network: networkIdentifier,
	}
	mockClient := &mocks.Client{}
	servicer := NewNetworkAPIService(cfg, mockClient)
	ctx := context.Background()

	networkList, err := servicer.NetworkList(ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, []*types.NetworkIdentifier{
		networkIdentifier,
	}, networkList.NetworkIdentifiers)

	networkStatus, err := servicer.NetworkStatus(ctx, nil)
	assert.Nil(t, networkStatus)
	assert.Equal(t, ErrUnavailableOffline.Code, err.Code)
	assert.Equal(t, ErrUnavailableOffline.Message, err.Message)

	networkOptions, err := servicer.NetworkOptions(ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, defaultNetworkOptions, networkOptions)

	mockClient.AssertExpectations(t)
}

func TestNetworkEndpoints_Online(t *testing.T) {
	cfg := &configuration.Configuration{
		Mode:                   configuration.Online,
		Network:                networkIdentifier,
		GenesisBlockIdentifier: wemix.MainnetGenesisBlockIdentifier,
	}
	mockClient := &mocks.Client{}
	servicer := NewNetworkAPIService(cfg, mockClient)
	ctx := context.Background()

	networkList, err := servicer.NetworkList(ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, []*types.NetworkIdentifier{
		networkIdentifier,
	}, networkList.NetworkIdentifiers)

	currentBlock := &types.BlockIdentifier{
		Index: 10,
		Hash:  "block 10",
	}

	currentTime := int64(1000000000000)

	syncStatus := &types.SyncStatus{
		CurrentIndex: types.Int64(100),
	}

	peers := []*types.Peer{
		{
			PeerID: "77.93.223.9:8333",
		},
	}

	mockClient.On(
		"Status",
		ctx,
	).Return(
		currentBlock,
		currentTime,
		syncStatus,
		peers,
		nil,
	)
	networkStatus, err := servicer.NetworkStatus(ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, &types.NetworkStatusResponse{
		GenesisBlockIdentifier: wemix.MainnetGenesisBlockIdentifier,
		CurrentBlockIdentifier: currentBlock,
		CurrentBlockTimestamp:  currentTime,
		Peers:                  peers,
		SyncStatus:             syncStatus,
	}, networkStatus)

	networkOptions, err := servicer.NetworkOptions(ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, defaultNetworkOptions, networkOptions)

	mockClient.AssertExpectations(t)
}
