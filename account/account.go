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

package account

import (
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/primasio/primas-node/config"
	"errors"
)

var nodeAccount *accounts.Account
var nodeKeystore *keystore.KeyStore

func Init () error {

	c := config.GetConfig()

	nodeKeystore = keystore.NewKeyStore(c.GetString("node_account.keystore_dir"), keystore.LightScryptN, keystore.LightScryptP)

	if len(nodeKeystore.Accounts()) == 0 {
		return errors.New("node account not found")
	}

	nodeAccount = &nodeKeystore.Accounts()[0]

	return nodeKeystore.Unlock(*nodeAccount, c.GetString("node_account.passphrase"))
}

func GetNodeAccount() *accounts.Account {
	return nodeAccount
}

func GetNodeKeystore() *keystore.KeyStore {
	return nodeKeystore
}