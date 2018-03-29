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
	"github.com/primasio/primas-node/incentives"
	"github.com/shopspring/decimal"
	"time"
)

var tokenContract *TokenContract = nil

type TokenContract struct {
	Contract *Contract
}

func GetTokenContract () (*TokenContract, error) {
	if tokenContract == nil {
		contract := new(TokenContract)

		var err error

		contract.Contract, err = GetContractByName("token")

		if err != nil {
			return nil, err
		}

		tokenContract = contract
	}

	return tokenContract, nil
}

func (tokenContract *TokenContract) Inflate() error {

	txHash, err := tokenContract.Contract.Execute("inflate")

	if err != nil {
		return err
	}

	log.Println("transaction hash: " + txHash)

	return nil
}

func (tokenContract *TokenContract) HandleEvent(eventLog *types.Log, db *gorm.DB) error {

	// We only need event name topic

	topic := eventLog.Topics[0]

	name, err := tokenContract.Contract.GetEventNameByTopicHash(topic.Hex())

	if err != nil {
		return err
	}

	log.Println("event triggered: " + name)

	switch name {
		case "Transfer":
			return tokenContract.handleTransfer(name, eventLog, db)
		case "Inflate":
			return tokenContract.handleInflate(name, eventLog, db)
		case "Lock":
			return tokenContract.handleLock(name, eventLog, db)
		default:
			return errors.New("unrecognized event")
	}

	return nil
}

func (tokenContract *TokenContract) handleTransfer(name string, eventLog *types.Log, db *gorm.DB) error {

	from := common.BytesToAddress(eventLog.Topics[1].Bytes())
	to := common.BytesToAddress(eventLog.Topics[2].Bytes())

	value := big.NewInt(0)
	value.SetBytes(eventLog.Data)

	// Update from
	if from.Big().Cmp(big.NewInt(0)) != 0 {
		tokenContract.updateUserBalance(from.Hex(), value, db, false)
	}

	// Update to
	if to.Big().Cmp(big.NewInt(0)) != 0 {
		tokenContract.updateUserBalance(to.Hex(), value, db, true)
	}

	return nil
}

func (tokenContract *TokenContract) handleInflate(name string, eventLog *types.Log, db *gorm.DB) error {

	amount := new(big.Int)

	err := tokenContract.Contract.ABI.Unpack(&amount, name, eventLog.Data)

	if err != nil {
		return err
	}

	incentives.DistributeIncentives(amount, db)

	incentiveContract, err := GetIncentiveContract()

	if err != nil {
		return err
	}

	return incentiveContract.AssignIncentives(db)
}

type TokenLockArgs struct {
	UserAddress common.Address
	ResourceType  *big.Int
	ResourceDNA []byte
	Amount      *big.Int
	Expire      *big.Int
}

func (tokenContract *TokenContract) handleLock(name string, eventLog *types.Log, db *gorm.DB) error {
	args := &TokenLockArgs{}

	err := tokenContract.Contract.ABI.Unpack(args, name, eventLog.Data)

	if err != nil {
		return err
	}

	tokenLock := &models.TokenLock{}

	tokenLock.UserAddress = args.UserAddress.Hex()
	tokenLock.ResourceType = uint(args.ResourceType.Uint64())
	tokenLock.ResourceDNA = string(args.ResourceDNA)

	tokenLock.Amount = decimal.NewFromBigInt(args.Amount, 0)
	tokenLock.Expire = uint(args.Expire.Uint64())
	tokenLock.CreatedAt = uint(time.Now().Unix())

	db.Save(tokenLock)

	return nil
}

func (tokenContract *TokenContract) updateUserBalance(address string, amount *big.Int, db *gorm.DB, isAdd bool) {
	user := &models.User{ Address: address }
	models.IdentifyUser(user, db)
	lockedUser := &models.User{ Address: user.Address }
	db.Set("gorm:query_option", "FOR UPDATE").Where(lockedUser).First(lockedUser)

	amountDecimal := decimal.NewFromBigInt(amount, 0)

	if isAdd {
		lockedUser.Balance = lockedUser.Balance.Add(amountDecimal)
	} else {
		lockedUser.Balance = lockedUser.Balance.Sub(amountDecimal)
	}

	db.Save(lockedUser)
}
