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
)

type System struct {
	gorm.Model
	Key string `gorm:"size:64;unique_index"`
	Value string `gorm:"type:text"`
}

func GetState (key string, dbi *gorm.DB) string {
	item := new(System)
	dbi.Where(&System{Key: key}).First(&item)
	return item.Value
}

func SetState (key, value string, dbi *gorm.DB) {
	item := new(System)
	dbi.Where(&System{Key: key}).First(&item)

	if item.Key == "" {
		// First time creation
		item.Key = key
	}

	item.Value = value
	dbi.Save(&item)
}
