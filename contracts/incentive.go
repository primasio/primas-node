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
	"math/big"
	"github.com/primasio/primas-node/models"
	"github.com/jinzhu/gorm"
)

var incentiveContract *IncentiveContract = nil

type IncentiveContract struct {
	Contract *Contract
}

func GetIncentiveContract () (*IncentiveContract, error) {
	if incentiveContract == nil {
		contract := new(IncentiveContract)

		var err error

		contract.Contract, err = GetContractByName("incentives")

		if err != nil {
			return nil, err
		}

		incentiveContract = contract
	}

	return incentiveContract, nil
}

func (incentiveContract *IncentiveContract) AssignIncentives (db *gorm.DB) error {
	batchSize := 200
	currentBatchOffset := 0

	userIncentives := make(map[string]*big.Int)

	for {
		var incs [] models.Incentive

		in := db.Table("incentives").Where("status = ?", models.IncentivesCalculating)
		in = in.Order("created_at asc").Offset(currentBatchOffset).Limit(batchSize)
		in.Find(&incs)

		if len(incs) == 0 {
			break
		}

		for _, incentive := range incs {

			userAddress := incentive.UserAddress
			amount := &big.Int{}
			amount.Set(incentive.Amount.Coefficient())

			if userIncentives[userAddress] != nil {
				userIncentives[userAddress].Add(userIncentives[userAddress], amount)
			} else {
				userIncentives[userAddress] = amount

				if len(userIncentives) == 100 {

					// Call contract to distribute
					if err := incentiveContract.Distribute(userIncentives); err != nil {
						return err
					}

					// Clear distributed
					userIncentives = make(map[string]*big.Int)
				}
			}
		}

		currentBatchOffset = currentBatchOffset + batchSize
	}

	if len(userIncentives) != 0 {
		if err := incentiveContract.Distribute(userIncentives); err != nil {
			return err
		}
	}

	// Finish calculation

	in2 := db.Table("incentives").Where("status = ?", models.IncentivesCalculating)
	in2.Updates(map[string]interface{}{"status": models.IncentivesPaid})

	return nil
}

func (incentiveContract *IncentiveContract) Distribute(incentives map[string]*big.Int) error {

	//var userAddresses []common.Address
	//var userAmounts []*big.Int
	//
	//for address, amount := range incentives {
	//	userAddresses = append(userAddresses, common.HexToAddress(address))
	//	userAmounts = append(userAmounts, amount)
	//}
	//
	//txHash, err := incentiveContract.Contract.Execute(
	//	"grantIncentives",
	//	userAddresses,
	//	userAmounts )
	//
	//if err != nil {
	//	return err
	//}
	//
	//log.Println("transaction hash: " + txHash)

	return nil
}
