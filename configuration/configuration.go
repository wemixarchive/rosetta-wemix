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

package configuration

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/wemixarchive/rosetta-wemix/wemix"
)

// Mode is the setting that determines if
// the implementation is "online" or "offline".
type Mode string

const (
	// Online is when the implementation is permitted
	// to make outbound connections.
	Online Mode = "ONLINE"

	// Offline is when the implementation is not permitted
	// to make outbound connections.
	Offline Mode = "OFFLINE"

	// Mainnet is the Wemix Mainnet.
	Mainnet string = "MAINNET"

	// Testnet is the Wemix Mainnet.
	Testnet string = "TESTNET"

	// DataDirectory is the default location for all
	// persistent data.
	DataDirectory = "/data"

	// ModeEnv is the environment variable read
	// to determine mode.
	ModeEnv = "MODE"

	// NetworkEnv is the environment variable
	// read to determine network.
	NetworkEnv = "NETWORK"

	// PortEnv is the environment variable
	// read to determine the port for the Rosetta
	// implementation.
	PortEnv = "PORT"

	// GwemixEnv is an optional environment variable
	// used to connect rosetta-wemix to an already
	// running gwemix node.
	GwemixEnv = "GWEMIX"

	// DefaultGwemixURL is the default URL for
	// a running gwemix node. This is used
	// when GwemixEnv is not populated.
	DefaultGwemixURL = "http://localhost:8588"

	// SkipGwemixAdminEnv is an optional environment variable
	// to skip gwemix `admin` calls which are typically not supported
	// by hosted node services. When not set, defaults to false.
	SkipGwemixAdminEnv = "SKIP_GWEMIX_ADMIN"

	// MiddlewareVersion is the version of rosetta-wemix.
	MiddlewareVersion = "0.0.4"
)

// Configuration determines how
type Configuration struct {
	Mode                   Mode
	Network                *types.NetworkIdentifier
	GenesisBlockIdentifier *types.BlockIdentifier
	GwemixURL              string
	RemoteGwemix           bool
	Port                   int
	GwemixArguments        string
	SkipGwemixAdmin        bool

	// Block Reward Data
	Params *params.ChainConfig
}

// LoadConfiguration attempts to create a new Configuration
// using the ENVs in the environment.
func LoadConfiguration() (*Configuration, error) {
	config := &Configuration{}

	modeValue := Mode(os.Getenv(ModeEnv))
	switch modeValue {
	case Online:
		config.Mode = Online
	case Offline:
		config.Mode = Offline
	case "":
		return nil, errors.New("MODE must be populated")
	default:
		return nil, fmt.Errorf("%s is not a valid mode", modeValue)
	}

	networkValue := os.Getenv(NetworkEnv)
	switch networkValue {
	case Mainnet:
		config.Network = &types.NetworkIdentifier{
			Blockchain: wemix.Blockchain,
			Network:    wemix.MainnetNetwork,
		}
		config.GenesisBlockIdentifier = wemix.MainnetGenesisBlockIdentifier
		config.Params = params.WemixMainnetChainConfig
		config.GwemixArguments = wemix.MainnetGwemixArguments
	case Testnet:
		config.Network = &types.NetworkIdentifier{
			Blockchain: wemix.Blockchain,
			Network:    wemix.TestnetNetwork,
		}
		config.GenesisBlockIdentifier = wemix.TestnetGenesisBlockIdentifier
		config.Params = params.WemixTestnetChainConfig
		config.GwemixArguments = wemix.TestnetGwemixArguments
	case "":
		return nil, errors.New("NETWORK must be populated")
	default:
		return nil, fmt.Errorf("%s is not a valid network", networkValue)
	}

	config.GwemixURL = DefaultGwemixURL
	envGwemixURL := os.Getenv(GwemixEnv)
	if len(envGwemixURL) > 0 {
		config.RemoteGwemix = true
		config.GwemixURL = envGwemixURL
	}

	config.SkipGwemixAdmin = false
	envSkipGwemixAdmin := os.Getenv(SkipGwemixAdminEnv)
	if len(envSkipGwemixAdmin) > 0 {
		val, err := strconv.ParseBool(envSkipGwemixAdmin)
		if err != nil {
			return nil, fmt.Errorf("%w: unable to parse SKIP_GWEMIX_ADMIN %s", err, envSkipGwemixAdmin)
		}
		config.SkipGwemixAdmin = val
	}

	portValue := os.Getenv(PortEnv)
	if len(portValue) == 0 {
		return nil, errors.New("PORT must be populated")
	}

	port, err := strconv.Atoi(portValue)
	if err != nil || len(portValue) == 0 || port <= 0 {
		return nil, fmt.Errorf("%w: unable to parse port %s", err, portValue)
	}
	config.Port = port

	return config, nil
}
