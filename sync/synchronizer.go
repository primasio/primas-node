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
	"github.com/primasio/primas-node/config"
	"context"
	"time"
	"log"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
	"github.com/primasio/primas-node/models"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/primasio/primas-node/db"
	"github.com/primasio/primas-node/contracts"
)

type Block struct {
	Number string
	Hash string
}

type BlockSynchronizer struct {
	ethClient *rpc.Client
	eventDispatcher *Dispatcher
	filter * ethereum.FilterQuery
}

func StartBlockSynchronizer () error {
	s := new(BlockSynchronizer)
	err := s.InitEthClient()

	if err != nil {
		return err
	}

	s.Start()

	return nil
}

func (synchronizer *BlockSynchronizer) Start() {

	blockChannel := make(chan Block)

	synchronizer.eventDispatcher = &Dispatcher{}
	synchronizer.eventDispatcher.Init()

	allContracts := contracts.GetAllContracts()

	if len(allContracts) == 0 {
		log.Println("Registered contract not found")
		return
	}

	// Filter
	synchronizer.filter = new(ethereum.FilterQuery)
	synchronizer.filter.Addresses = [] common.Address{}
	synchronizer.filter.Topics = [][] common.Hash {}

	for _, ctr := range allContracts {
		synchronizer.filter.Addresses = append(synchronizer.filter.Addresses, ctr.Address)
	}

	// Subscribe to block updating

	go func() {
		for i := 0; ; i++ {
			if i > 0 {
				time.Sleep(2 * time.Second)
			}

			err := synchronizer.subscribeBlocks(blockChannel)

			if err != nil {
				log.Println("new block subscription failed: ", err)
			}
		}
	}()

	// Synchronize to new blocks as they arrive.
	for block := range blockChannel {

		n := new(big.Int)

		n, ok := n.SetString(block.Number[2:], 16)

		log.Println("new block: #" + n.String())

		if !ok {
			log.Println("invalid block number")
		} else {

			// Update latest block hash

			models.SetState("CurrentBlockHash", block.Hash, db.GetDb())

			err := synchronizer.syncTo(n.Sub(n, big.NewInt(6)))

			if err != nil {
				log.Println("block synchronization failed: ", err)
			}
		}
	}
}

func (synchronizer *BlockSynchronizer) InitEthClient () error {
	if synchronizer.ethClient == nil {
		c := config.GetConfig()

		clt, err := rpc.Dial(c.GetString("eth_node.protocol") + "://" + c.GetString("eth_node.host") + ":" + c.GetString("eth_node.port"))

		if err != nil {
			return err
		}

		synchronizer.ethClient = clt
	}

	return nil
}

func (synchronizer *BlockSynchronizer) subscribeBlocks(block chan Block) error {

	c := config.GetConfig()

	duration, err := time.ParseDuration(c.GetString("eth_node.timeout"))

	if err != nil {
		return err
	}

	ctx, _ := context.WithTimeout(context.Background(), duration)

	// Subscribe to new blocks.
	sub, err := synchronizer.ethClient.EthSubscribe(ctx, block, "newHeads")

	if err != nil {
		return err
	}

	// The subscription will deliver events to the channel. Wait for the
	// subscription to end for any reason, then loop around to re-establish
	// the connection.

	log.Println("connection lost: ", <-sub.Err())
	return nil
}

var synchronizing = false

func (synchronizer *BlockSynchronizer) syncTo(toBlockNumber *big.Int) error {

	if synchronizing {
		return nil
	}

	synchronizing = true

	currentBlockNumber := big.NewInt(0)
	var ok bool

	currentBlockNumberStr := models.GetState("CurrentBlockNumber", db.GetDb())

	if currentBlockNumberStr != "" {

		currentBlockNumber, ok = currentBlockNumber.SetString(currentBlockNumberStr, 10)

		if !ok {
			synchronizing = false
			return errors.New("invalid current block number")
		}

	} else {
		c := config.GetConfig()
		currentBlockNumber.SetInt64(c.GetInt64("synchronizer.start_block"))
	}

	batchSize := big.NewInt(100000)

	for
	;
	currentBlockNumber.Cmp(toBlockNumber) <= 0;
	currentBlockNumber.Add(currentBlockNumber, batchSize) {

		start := new(big.Int)
		start.Set(currentBlockNumber)
		start.Add(start, big.NewInt(1))


		end := new(big.Int)
		end.Set(currentBlockNumber)
		end.Add(end, batchSize)

		if end.Cmp(toBlockNumber) > 0 {
			end.Set(toBlockNumber)
		}

		err := synchronizer.syncRange(start, end)

		if err != nil {
			synchronizing = false
			return err
		}
	}

	synchronizing = false
	return nil
}

func (synchronizer *BlockSynchronizer) syncRange(start, end *big.Int) error {

	c := config.GetConfig()

	// Context
	duration, err2 := time.ParseDuration(c.GetString("eth_node.timeout"))

	if err2 != nil {
		return err2
	}

	// Filter Range
	synchronizer.filter.FromBlock = start
	synchronizer.filter.ToBlock = end

	var logItems []types.Log

	ctx, _ := context.WithTimeout(context.Background(), duration)

	err := synchronizer.ethClient.CallContext(ctx, &logItems, "eth_getLogs", toFilterArg(*synchronizer.filter))

	if err != nil {
		return err
	}

	// Process log items in a transaction

	// Start transaction
	tx := db.GetDb().Begin()

	for _, logItem := range logItems {

		err := synchronizer.eventDispatcher.DispatchEvent(&logItem, tx)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	models.SetState("CurrentBlockNumber", end.String(), tx)

	// Commit transaction
	tx.Commit()

	log.Println("synchronized to block #" + end.String())

	return nil
}

func toFilterArg(q ethereum.FilterQuery) interface{} {
	arg := map[string]interface{}{
		"fromBlock": toBlockNumArg(q.FromBlock),
		"toBlock":   toBlockNumArg(q.ToBlock),
		"address":   q.Addresses,
		"topics":    q.Topics,
	}
	if q.FromBlock == nil {
		arg["fromBlock"] = "0x0"
	}
	return arg
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	return hexutil.EncodeBig(number)
}