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
	"github.com/jinzhu/gorm"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"github.com/primasio/primas-node/models"
	"encoding/hex"
)

var userContract *UserContract = nil

type UserContract struct {
	Contract *Contract
}

func GetUserContract () (*UserContract, error) {
	if userContract == nil {
		contract := new(UserContract)

		var err error

		contract.Contract, err = GetContractByName("user")

		if err != nil {
			return nil, err
		}

		userContract = contract
	}

	return userContract, nil
}

func (userContract *UserContract) Burn (timestamp, userAddress, signature string) error {

	sigBytes, err := hex.DecodeString(signature)

	if err != nil {
		return err
	}

	address := common.HexToAddress(userAddress)

	if err != nil {
		return err
	}

	txHash, err := userContract.Contract.Execute(
		"burn",
		timestamp,
		sigBytes,
		address )

	if err != nil {
		return err
	}

	log.Println("transaction hash: " + txHash)

	return nil
}

type UserTokenBurnArgs struct {
	UserAddress common.Address
	Amount *big.Int
}

func (userContract *UserContract) HandleEvent(eventLog *types.Log, db *gorm.DB) error {

	// We only need event name topic

	topic := eventLog.Topics[0]

	name, err := userContract.Contract.GetEventNameByTopicHash(topic.Hex())

	if err != nil {
		return err
	}

	log.Println("event triggered: " + name)

	switch name {
		case "UserTokenBurnLog":
			return userContract.handleUserTokenBurn(name, eventLog, db)
		default:
			return errors.New("unrecognized event")
	}

	return nil
}

func (userContract *UserContract) handleUserTokenBurn(name string, eventLog *types.Log, db *gorm.DB) error {
	args := &UserTokenBurnArgs{}

	err := userContract.Contract.ABI.Unpack(args, name, eventLog.Data)

	if err != nil {
		return err
	}

	user := &models.User{ Address: args.UserAddress.Hex() }

	db.Set("gorm:query_option", "FOR UPDATE").Where(user).First(&user)

	if user.ID == 0 {
		models.IdentifyUser(user, db)
	}

	user.TokenBurned = 1
	db.Save(user)

	return nil
}
