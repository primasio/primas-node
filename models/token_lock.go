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

import "github.com/shopspring/decimal"

const TokenLockResourceGroup = 0
const TokenLockResourceArticle = 1

type TokenLock struct {
	ID              uint `gorm:"primary_key"`
	CreatedAt       uint
	UserAddress     string `gorm:"size:255;index"`
	ResourceType    uint `gorm:"index"`
	ResourceDNA     string `gorm:"index"`
	Amount          decimal.Decimal `gorm:"type:decimal(65)"`
	Expire          uint `gorm:"type:int unsigned;index"`
}