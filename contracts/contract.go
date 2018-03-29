/*
 * Copyright 2017 Primas Lab Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package contracts

import (
	"strings"
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/primasio/primas-node/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/primasio/primas-node/account"
	"math/big"
	"time"
	"errors"
	"sync"
)

type Contract struct {
	Address common.Address
	ABI abi.ABI
	eventNameHashMap map[string]string
}

var contractsByName map[string]*Contract
var currentNonce uint64
var contractMutex *sync.Mutex

func InitContracts () error {

	contractsByName = make(map[string]*Contract)

	c := config.GetConfig()

	cMap := c.GetStringMap("contracts")

	for name, item := range cMap {

		itemMap, ok := item.(map[string]interface{})

		if !ok {
			return errors.New("contract initialization failed")
		}

		address, ok := itemMap["address"].(string)

		if !ok {
			return errors.New("contract address initialization failed")
		}

		abiJson, ok := itemMap["abi"].(string)

		if !ok {
			return errors.New("contract abi initialization failed")
		}

		nContract, err := NewContract(address, abiJson)

		if err != nil {
			return err
		}

		contractsByName[name] = nContract
	}

	contractMutex = &sync.Mutex{}
	currentNonce = 0

	return nil
}

func GetContractByName(name string) (*Contract, error) {
	if contractsByName[name] == nil {
		return nil, errors.New("contract " + name + " does not exist")
	}

	return contractsByName[name], nil
}

func GetAllContracts() map[string]*Contract {
	return contractsByName
}

func NewContract(address, abi string) (*Contract, error) {

	if address == "" || abi == "" {
		return nil, errors.New("contract address and abi cannot be blank")
	}

	contract := new(Contract)

	contract.Address = common.HexToAddress(address)

	err := contract.InitABI(abi)

	if err != nil {
		return nil, err
	}

	contract.InitEventNameHashMap()

	return contract, nil
}

func (contract *Contract) InitABI (ABIJson string) error {
	abiInstance, err := abi.JSON(strings.NewReader(ABIJson))

	if err != nil {
		return err
	}

	contract.ABI = abiInstance

	return nil
}

func (contract *Contract) Execute (method string, args ...interface{}) (string, error) {

	c := config.GetConfig()

	methodBytes, err := contract.ABI.Pack(method, args...)

	if err != nil {
		return "", err
	}

	client, err := contract.GetEthClient()

	if err != nil {
		return "", err
	}

	nodeAccount := account.GetNodeAccount()

	duration, err2 := time.ParseDuration(c.GetString("eth_node.timeout"))

	if err2 != nil {
		return "", err2
	}

	contractMutex.Lock()

	ctx, _ := context.WithTimeout(context.Background(), duration)

	nonce, err := ethClient.NonceAt(ctx, nodeAccount.Address, nil)

	if err != nil {
		return "", err
	}

	if nonce > currentNonce || currentNonce == 0 {
		currentNonce = nonce
	} else {
		currentNonce = currentNonce + 1
	}

	tx := types.NewTransaction(
		currentNonce,
		contract.Address,
		big.NewInt(0),
		big.NewInt(c.GetInt64("node_account.gas_limit")),
		big.NewInt(c.GetInt64("node_account.gas_price")),
		methodBytes)

	ks := account.GetNodeKeystore()
	signedTx, err := ks.SignTx(*nodeAccount, tx, big.NewInt(c.GetInt64("eth_node.chain_id")))

	if err != nil {
		contractMutex.Unlock()
		return "", err
	}

	ctx2, _ := context.WithTimeout(context.Background(), duration)

	if err := client.SendTransaction(ctx2, signedTx); err != nil {
		contractMutex.Unlock()
		return "", err
	}

	contractMutex.Unlock()
	return signedTx.Hash().String(), nil
}

var ethClient *ethclient.Client

func (contract *Contract) GetEthClient () (*ethclient.Client, error) {
	if ethClient == nil {
		c := config.GetConfig()

		client, err := ethclient.Dial(c.GetString("eth_node.protocol") + "://" + c.GetString("eth_node.host") + ":" + c.GetString("eth_node.port"))

		if err != nil {
			return nil, err
		}

		ethClient = client
	}

	return ethClient, nil
}

func (contract *Contract) GetEventNameByTopicHash(hash string) (string, error) {
	if contract.eventNameHashMap[hash] == "" {
		return "", errors.New("topic does not exist")
	}

	return contract.eventNameHashMap[hash], nil
}

func (contract *Contract) InitEventNameHashMap() {

	contract.eventNameHashMap = make(map[string]string)

	for _, event := range contract.ABI.Events {
		contract.eventNameHashMap[event.Id().Hex()] = event.Name
	}
}