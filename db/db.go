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

package db

import (
	"github.com/primasio/primas-node/config"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"os"
)

var instance *gorm.DB

func GetDb() *gorm.DB {
	return instance
}

func Init() error {

	c := config.GetConfig()

	dbType := c.GetString("db.type")
	dbConn := c.GetString("db.connection")

	if dbConn == "" {
		f, err := ioutil.TempFile("", "")
		if err != nil {
			panic(err)
		}
		dbConn := f.Name()
		f.Close()
		os.Remove(dbConn)
	}

	var err error

	instance, err = gorm.Open(dbType, dbConn)

	if err != nil {
		return err
	}

	return nil
}
