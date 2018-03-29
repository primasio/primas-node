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
	"log"
	"github.com/primasio/primas-node/models"
	"github.com/ethereum/go-ethereum/common"
	"encoding/hex"
)

var contentContract *ContentContract = nil

type ContentContract struct {
	Contract *Contract
}

type Content interface {
	ToMetadata () (title, contentHash, license, blockHash, extras, signature, DNA []byte, err error)
	FromMetadata (title, contentHash, license, blockHash, extras, signature, DNA []byte) error
	GetUserAddress () string
	GetDNA () string
}


func GetContentContract() (*ContentContract, error) {

	if contentContract == nil {
		contract := new(ContentContract)

		var err error

		contract.Contract, err = GetContractByName("content")

		if err != nil {
			return nil, err
		}

		contentContract = contract
	}

	return contentContract, nil
}

func (contentContract *ContentContract) Publish (content Content) error {

	title, contentHash, license, blockHash, extras, signature, DNA, err := content.ToMetadata()

	if err != nil {
		return err
	}

	address := common.HexToAddress(content.GetUserAddress())

	txHash, err := contentContract.Contract.Execute(
		"publish",
		title,
		contentHash,
		license,
		extras,
		blockHash,
		signature,
		DNA,
		address)

	if err != nil {
		return err
	}

	log.Println("transaction hash: " + txHash)

	return nil
}

func (contentContract *ContentContract) Like (like *models.ArticleLike) error {

	sigBytes, err := hex.DecodeString(like.Signature)

	if err != nil {
		return err
	}

	address := common.HexToAddress(like.GroupMemberAddress)

	txHash, err := contentContract.Contract.Execute(
		"like",
		[]byte(like.ArticleDNA),
		[]byte(like.GroupDNA),
		sigBytes,
		address )

	if err != nil {
		return err
	}

	log.Println("transaction hash: " + txHash)

	return nil
}

func (contentContract *ContentContract) Comment (comment *models.ArticleComment) error {

	sigBytes, err := hex.DecodeString(comment.Signature)

	if err != nil {
		return err
	}

	address := common.HexToAddress(comment.GroupMemberAddress)

	txHash, err := contentContract.Contract.Execute(
		"comment",
		[]byte(comment.ArticleDNA),
		[]byte(comment.GroupDNA),
		[]byte(comment.ContentHash),
		sigBytes,
		address)

	if err != nil {
		return err
	}

	log.Println("transaction hash: " + txHash)

	return nil
}

func (contentContract *ContentContract) Share (share *models.ArticleShareBatch) error {

	sigBytes, err := hex.DecodeString(share.Signature)

	if err != nil {
		return err
	}

	groupsDNA := share.GetConcatenatedGroupDNAString()

	address := common.HexToAddress(share.GroupMemberAddress)

	txHash, err := contentContract.Contract.Execute(
		"share",
		[]byte(share.ArticleDNA),
		[]byte(groupsDNA),
		sigBytes,
		address )

	if err != nil {
		return err
	}

	log.Println("transaction hash: " + txHash)

	return nil
}

