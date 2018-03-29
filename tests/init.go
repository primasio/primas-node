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

package tests

import (
	"flag"
	"os"
	"github.com/primasio/primas-node/config"
	"github.com/primasio/primas-node/db"
	"github.com/primasio/primas-node/models"
	"github.com/primasio/primas-node/contracts"
	"github.com/primasio/primas-node/account"
	"log"
	"time"
	"math/rand"
)

func InitTestEnv(configPath string) {
	environment := flag.String("e", "test", "")

	flag.Parse()

	// Init Config
	config.Init(*environment, &configPath)

	// Init Database
	if err := db.Init(); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// Update Database Models
	models.MigrateAll()

	// Init Contracts
	if err := contracts.InitContracts(); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// Init Node Account
	if err := account.Init(); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	rand.Seed(time.Now().UnixNano())
}
