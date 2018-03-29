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

package sync

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jinzhu/gorm"
	"errors"
	"github.com/primasio/primas-node/contracts"
)

type LogEventHandler interface {
	HandleEvent(eventLog *types.Log, db *gorm.DB) error
}

type Dispatcher struct {}

var eventHandlerRegistry map[string]LogEventHandler

func (dispatcher *Dispatcher) Init () error {

	// Event handler registry
	// All handlers must be registered here

	eventHandlerRegistry = make(map[string]LogEventHandler)

	if contract, err := contracts.GetMetadataContract(); err == nil {
		eventHandlerRegistry[contract.Contract.Address.Hex()] = contract
	}else {
		return err
	}

	if contract, err := contracts.GetGroupContract(); err == nil {
		eventHandlerRegistry[contract.Contract.Address.Hex()] = contract
	}else {
		return err
	}

	if contract, err := contracts.GetTokenContract(); err == nil {
		eventHandlerRegistry[contract.Contract.Address.Hex()] = contract
	}else {
		return err
	}

	if contract, err := contracts.GetUserContract(); err == nil {
		eventHandlerRegistry[contract.Contract.Address.Hex()] = contract
	}else {
		return err
	}

	return nil
}

func (dispatcher *Dispatcher) DispatchEvent (log *types.Log, db *gorm.DB) error {

	addr := log.Address.String()

	if eventHandlerRegistry[addr] == nil {
		return errors.New("log event handler does not exist: " + addr)
	}

	return eventHandlerRegistry[addr].HandleEvent(log, db)
}