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
	"github.com/primasio/primas-node/models"
	"github.com/jinzhu/gorm"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"errors"
	"encoding/hex"
	"time"
	"github.com/primasio/primas-node/crypto"
	"github.com/ethereum/go-ethereum/common"
)

var groupContract *GroupContract = nil

type GroupContract struct {
	Contract *Contract
}

func GetGroupContract() (*GroupContract, error) {

	if groupContract == nil {
		contract := new(GroupContract)

		var err error

		contract.Contract, err = GetContractByName("group")

		if err != nil {
			return nil, err
		}

		groupContract = contract
	}

	return groupContract, nil
}

func (groupContract *GroupContract) Create(group *models.Group) error {

	sigBytes, err := hex.DecodeString(group.Signature)

	if err != nil {
		return err
	}

	address := common.HexToAddress(group.UserAddress)

	txHash, err := groupContract.Contract.Execute(
		"create",
		[]byte(group.DNA),
		[]byte(group.Title),
		[]byte(group.Description),
		sigBytes,
		address )

	if err != nil {
		return err
	}

	log.Println("transaction hash: " + txHash)

	return nil
}

func (groupContract *GroupContract) AddMember(member *models.GroupMember) error {
	sigBytes, err := hex.DecodeString(member.Signature)

	if err != nil {
		return err
	}

	address := common.HexToAddress(member.MemberAddress)

	txHash, err := groupContract.Contract.Execute(
		"addMember",
		[]byte(member.GroupDNA),
		sigBytes,
		address)

	if err != nil {
		return err
	}

	log.Println("transaction hash: " + txHash)

	return nil
}

func (groupContract *GroupContract) RemoveMember(member *models.GroupMember) error {
	sigBytes, err := hex.DecodeString(member.Signature)

	if err != nil {
		return err
	}

	address := common.HexToAddress(member.MemberAddress)

	txHash, err := groupContract.Contract.Execute(
		"removeMember",
		[]byte(member.GroupDNA),
		sigBytes,
		address )

	if err != nil {
		return err
	}

	log.Println("transaction hash: " + txHash)

	return nil
}

func (groupContract *GroupContract) RemoveMemberByOwner(member *models.GroupMember, ownerAddress string) error {
	sigBytes, err := hex.DecodeString(member.Signature)

	if err != nil {
		return err
	}

	address := common.HexToAddress(ownerAddress)

	txHash, err := groupContract.Contract.Execute(
		"removeMemberByOwner",
		[]byte(member.GroupDNA),
		[]byte(member.MemberAddress),
		sigBytes,
		address )

	if err != nil {
		return err
	}

	log.Println("transaction hash: " + txHash)

	return nil
}

type CreateLogArgs struct {
	Title                []byte
	Description          []byte
	Signature            []byte
}

type AddMemberLogArgs struct {
	GroupDNA             []byte
	Signature            []byte
}

type RemoveMemberLogArgs struct {
	GroupDNA             []byte
	Signature            []byte
}

type RemoveMemberByOwnerLogArgs struct {
	GroupDNA             []byte
	GroupMemberAddress   []byte
	Signature            []byte
}

func (groupContract *GroupContract) HandleEvent(eventLog *types.Log, db *gorm.DB) error {

	for _, topic := range eventLog.Topics {

		name, err := groupContract.Contract.GetEventNameByTopicHash(topic.Hex())

		if err != nil {
			return err
		}

		log.Println("event triggered: " + name)

		switch name {
			case "CreateLog":
				return groupContract.handleCreate(name, eventLog, db)
			case "AddMemberLog":
				return groupContract.handleAddMember(name, eventLog, db)
			case "RemoveMemberLog":
				return groupContract.handleRemoveMember(name, eventLog, db)
			case "RemoveMemberByOwnerLog":
				return groupContract.handleRemoveMemberByOwner(name, eventLog, db)
			default:
				return errors.New("unrecognized event: " + name)
		}
	}

	return nil
}

func (groupContract *GroupContract) handleCreate(name string, eventLog *types.Log, db *gorm.DB) error {

	args := &CreateLogArgs{}

	err := groupContract.Contract.ABI.Unpack(args, name, eventLog.Data)

	if err != nil {
		return err
	}

	group := &models.Group{}

	group.Title = string(args.Title)
	group.Description = string(args.Description)

	sigBase := group.GetSignatureBaseString()
	signature := hex.EncodeToString(args.Signature)

	if user, err := crypto.ExtractUserFromSignature(sigBase, signature); err == nil {
		group.UserAddress = user.Address
	} else {
		return err
	}

	group.Signature = signature

	if dna, err := group.GenerateDNA(); err == nil {
		group.DNA = dna
	} else {
		return err
	}

	check := &models.Group{ DNA: group.DNA }

	db.Where(check).First(group)

	if group.ID == 0 {
		group.CreatedAt = uint(time.Now().Unix())
	}

	group.TxStatus = models.TxStatusConfirmed

	db.Save(group)

	groupMember := &models.GroupMember{ GroupDNA: group.DNA, MemberAddress: group.UserAddress }

	db.Where(groupMember).First(groupMember)

	if groupMember.ID == 0 {
		groupMember.CreatedAt = uint(time.Now().Unix())
	}

	groupMember.TxStatus = models.TxStatusConfirmed

	db.Save(groupMember)

	return nil
}

func (groupContract *GroupContract) handleAddMember(name string, eventLog *types.Log, db *gorm.DB) error {
	args := &AddMemberLogArgs{}

	err := groupContract.Contract.ABI.Unpack(args, name, eventLog.Data)

	if err != nil {
		return err
	}

	groupMember := &models.GroupMember{}

	groupMember.GroupDNA = string(args.GroupDNA)

	sigBase := groupMember.GetSignatureBaseString()
	signature := hex.EncodeToString(args.Signature)

	if user, err := crypto.ExtractUserFromSignature(sigBase, signature); err == nil {
		groupMember.MemberAddress = user.Address
	} else {
		return err
	}

	db.Where(groupMember).First(groupMember)

	if groupMember.ID == 0 {
		groupMember.CreatedAt = uint(time.Now().Unix())
	}

	groupMember.Signature = signature

	groupMember.TxStatus = models.TxStatusConfirmed

	db.Save(groupMember)

	group := &models.Group{ DNA: groupMember.GroupDNA }

	db.Set("gorm:query_option", "FOR UPDATE").Where(group).First(group)

	if group.ID == 0 {
		return errors.New("group does not exist")
	}

	group.MemberCount = group.MemberCount + 1

	db.Save(group)

	return nil
}

func (groupContract *GroupContract) handleRemoveMember(name string, eventLog *types.Log, db *gorm.DB) error {
	args := &RemoveMemberLogArgs{}

	err := groupContract.Contract.ABI.Unpack(args, name, eventLog.Data)

	if err != nil {
		return err
	}

	groupMember := &models.GroupMember{}

	groupMember.GroupDNA = string(args.GroupDNA)

	sigBase := groupMember.GetSignatureBaseString()
	signature := hex.EncodeToString(args.Signature)

	if user, err := crypto.ExtractUserFromSignature(sigBase, signature); err == nil {
		groupMember.MemberAddress = user.Address
	} else {
		return err
	}

	db.Where(groupMember).First(groupMember)

	if groupMember.ID == 0 {
		return errors.New("member is not in group")
	}

	db.Delete(groupMember)

	group := &models.Group{ DNA: groupMember.GroupDNA }

	db.Set("gorm:query_option", "FOR UPDATE").Where(group).First(group)

	if group.ID == 0 {
		return errors.New("group does not exist")
	}

	group.MemberCount = group.MemberCount - 1

	db.Save(group)

	return nil
}

func (groupContract *GroupContract) handleRemoveMemberByOwner(name string, eventLog *types.Log, db *gorm.DB) error {
	args := &RemoveMemberByOwnerLogArgs{}

	err := groupContract.Contract.ABI.Unpack(args, name, eventLog.Data)

	if err != nil {
		return err
	}

	groupMember := &models.GroupMember{}

	groupMember.GroupDNA = string(args.GroupDNA)
	groupMember.MemberAddress = string(args.GroupMemberAddress)

	db.Where(groupMember).First(groupMember)

	if groupMember.ID == 0 {
		return errors.New("member is not in group")
	}

	db.Delete(groupMember)

	group := &models.Group{ DNA: groupMember.GroupDNA }

	db.Set("gorm:query_option", "FOR UPDATE").Where(group).First(group)

	if group.ID == 0 {
		return errors.New("group does not exist")
	}

	group.MemberCount = group.MemberCount - 1

	db.Save(group)

	return nil
}
