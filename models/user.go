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

package models

import (
	"github.com/jinzhu/gorm"
	"time"
	"math/big"
	"github.com/shopspring/decimal"
)

type User struct {
	ID                 uint `gorm:"primary_key"`
	CreatedAt          uint
	Address            string `gorm:"size:255;unique_index" binding:"required"`
	Name               string `gorm:"type:text" binding:"required"`
	Extra              string `gorm:"type:text" binding:"required"`
	Signature          string `gorm:"type:text" binding:"required"`
	Balance            decimal.Decimal `gorm:"type:decimal(65)" binding:"-"`
	TokenBurned        int    `gorm:"type:tinyint" binding:"-"`

	// Relations

	UserArticles       []Article `sql:"-" binding:"-"`
	UserGroups         []Group `sql:"-" binding:"-"`
}

func (user *User) GetSpendableBalance(db *gorm.DB) *big.Int {

	recordedBalance := user.Balance.Coefficient()
	lockedBalance := user.GetLockedBalance(db)

	return recordedBalance.Sub(recordedBalance, lockedBalance)
}

func (user *User) GetLockedBalance(db *gorm.DB) *big.Int {

	var tokenLocks []TokenLock

	now := time.Now().Unix()

	in := db.Table("token_locks")
	in = in.Where(&TokenLock{UserAddress: user.Address})
	in = in.Where("( expire = 0 OR expire > ? )", now)
	in.Find(&tokenLocks)

	sum := big.NewInt(0)

	for _, l := range tokenLocks {

		am := l.Amount.Coefficient()
		sum = sum.Add(sum, am)
	}

	return sum
}

func (user *User) GetHP(db *gorm.DB) *big.Int {

	theta := int64( 5 )            // threshold
	windowSize := uint( 12 * 3600 ) // window size 12 hours

	thetaInner := big.NewInt(theta)
	thetaOuter := big.NewInt(theta)

	// Operation count
	operationCount := 0

	in := db.Table("incentives").Where(&Incentive{UserAddress: user.Address})
	in = in.Where("created_at > ?", uint(time.Now().Unix()) - windowSize)
	in.Count(&operationCount)

	balance := user.Balance.Coefficient()

	e := big.NewInt(3) // E = 2.7 = 3

	Cj := big.NewInt(int64(operationCount))

	exp := e.Exp(e, thetaInner.Sub(thetaInner, Cj), nil)

	expA1 := exp.Add(exp, big.NewInt(1))

	CjDivExpA1 :=  Cj.Div(Cj, expA1)

	lower := thetaOuter.Add(thetaOuter, CjDivExpA1)

	lower = lower.Mul(lower, lower)

	hp := balance.Div(balance, lower)

	tenE18 := big.NewInt(10)
	tenE18 = tenE18.Exp(tenE18, big.NewInt(18), nil)

	hp = hp.Div(hp, tenE18)

	return hp
}

func IdentifyUser (user *User, db *gorm.DB) {

	db.Where(user).First(&user)

	if user.ID == 0 {
		// New user

		user.Name = ""
		user.Extra = "{}"
		user.Signature = ""

		user.CreatedAt = uint(time.Now().Unix())

		user.Balance = decimal.Zero
		user.TokenBurned = 0

		db.Save(&user)
	}
}