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
	"github.com/jinzhu/gorm"
	"github.com/ethereum/go-ethereum/core/types"
	"errors"
	"github.com/primasio/primas-node/models"
	"time"
	"github.com/primasio/primas-node/crypto"
	"encoding/hex"
)

var metadataContract *MetadataContract = nil

type MetadataContract struct {
	Contract *Contract
}

func GetMetadataContract() (*MetadataContract, error) {

	if metadataContract == nil {
		contract := new(MetadataContract)

		var err error

		contract.Contract, err = GetContractByName("metadata")

		if err != nil {
			return nil, err
		}

		metadataContract = contract
	}

	return metadataContract, nil
}

type PublishLogArgs struct {
	Title               []byte
	ContentHash         []byte
	License             []byte
	Extras              []byte
	BlockHash           []byte
	Signature           []byte
	DNA                 []byte
}

type LikeLogArgs struct {
	ArticleDNA          []byte
	GroupDNA            []byte
	Signature           []byte
}

type CommentLogArgs struct {
	ArticleDNA          []byte
	GroupDNA            []byte
	ContentHash         []byte
	Signature           []byte
}

type ShareLogArgs struct {
	ArticleDNA          []byte
	GroupsDNA           []byte
	Signature           []byte
}

func (metadataContract *MetadataContract) HandleEvent(eventLog *types.Log, db *gorm.DB) error {

	// We only need event name topic

	topic := eventLog.Topics[0]

	name, err := metadataContract.Contract.GetEventNameByTopicHash(topic.Hex())

	if err != nil {
		return err
	}

	log.Println("event triggered: " + name)

	switch name {
	case "PublishLog":
		return metadataContract.handlePublish(name, eventLog, db)
	case "LikeLog":
		return metadataContract.handleLike(name, eventLog, db)
	case "CommentLog":
		return metadataContract.handleComment(name, eventLog, db)
	case "ShareLog":
		return metadataContract.handleShare(name, eventLog, db)
	default:
		return errors.New("unrecognized event")
	}

	return nil
}

func (metadataContract *MetadataContract) handlePublish(name string, eventLog *types.Log, db *gorm.DB) error {

	args := &PublishLogArgs{}

	err := metadataContract.Contract.ABI.Unpack(args, name, eventLog.Data)

	if err != nil {
		return err
	}

	// Now that we have only article...

	article := &models.Article{}

	err = article.FromMetadata(args.Title, args.ContentHash, args.License, args.BlockHash, args.Extras, args.Signature, args.DNA)

	if err != nil {
		return err
	}

	db.Where(&models.Article{ DNA: article.DNA}).First(article)

	if article.ID == 0 {

		// Article does not exists
		author, err := crypto.ExtractUserFromSignature(article.GetSignatureBaseString(), article.Signature)

		if err != nil {
			return err
		}

		models.IdentifyUser(author, db)

		article.UserAddress = author.Address
		article.CreatedAt = uint(time.Now().Unix())
	}

	article.TxStatus = models.TxStatusConfirmed

	db.Save(article)

	return nil
}

func (metadataContract *MetadataContract) handleLike(name string, eventLog *types.Log, db *gorm.DB) error {
	args := &LikeLogArgs{}

	err := metadataContract.Contract.ABI.Unpack(args, name, eventLog.Data)

	if err != nil {
		return err
	}

	like := &models.ArticleLike{}

	like.ArticleDNA = string(args.ArticleDNA)
	like.GroupDNA = string(args.GroupDNA)

	sigBase := like.GetSignatureBaseString()
	signature := hex.EncodeToString(args.Signature)

	if user, err := crypto.ExtractUserFromSignature(sigBase, signature); err == nil {

		like.GroupMemberAddress = user.Address

	} else {
		return err
	}

	db.Where(like).First(like)

	if like.ID == 0 {
		like.CreatedAt = uint(time.Now().Unix())
	}

	like.Signature = signature
	like.TxStatus = models.TxStatusConfirmed

	db.Save(like)

	article := &models.Article{}
	db.Set("gorm:query_option", "FOR UPDATE").Where(&models.Article{DNA: like.ArticleDNA}).First(article)

	article.LikeCount = article.LikeCount + 1

	db.Save(article)

	models.LikeArticleIncentive(like, db)

	return nil
}

func (metadataContract *MetadataContract) handleComment(name string, eventLog *types.Log, db *gorm.DB) error {
	args := &CommentLogArgs{}

	err := metadataContract.Contract.ABI.Unpack(args, name, eventLog.Data)

	if err != nil {
		return err
	}

	comment := &models.ArticleComment{}

	comment.ArticleDNA = string(args.ArticleDNA)
	comment.GroupDNA = string(args.GroupDNA)
	comment.ContentHash = string(args.ContentHash)

	sigBase := comment.GetSignatureBaseString()
	signature := hex.EncodeToString(args.Signature)

	if user, err := crypto.ExtractUserFromSignature(sigBase, signature); err == nil {

		comment.GroupMemberAddress = user.Address

	} else {
		return err
	}

	db.Where(comment).First(comment)

	if comment.ID == 0 {
		comment.CreatedAt = uint(time.Now().Unix())
	}

	comment.Signature = signature
	comment.TxStatus = models.TxStatusConfirmed

	db.Set("gorm:save_associations", false).Save(comment)

	article := &models.Article{}
	db.Set("gorm:query_option", "FOR UPDATE").Where(&models.Article{DNA: comment.ArticleDNA}).First(article)

	article.CommentCount = article.CommentCount + 1

	db.Save(article)

	models.CommentArticleIncentive(comment, db)

	return nil
}

func (metadataContract *MetadataContract) handleShare(name string, eventLog *types.Log, db *gorm.DB) error {
	args := &ShareLogArgs{}

	err := metadataContract.Contract.ABI.Unpack(args, name, eventLog.Data)

	if err != nil {
		return err
	}

	shareBatch := &models.ArticleShareBatch{}

	shareBatch.ArticleDNA = string(args.ArticleDNA)
	shareBatch.FromConcatenatedGroupDNA(string(args.GroupsDNA))

	sigBase := shareBatch.GetSignatureBaseString()
	signature := hex.EncodeToString(args.Signature)

	if user, err := crypto.ExtractUserFromSignature(sigBase, signature); err == nil {

		shareBatch.GroupMemberAddress = user.Address

	} else {
		return err
	}

	for _, groupDNA := range shareBatch.GroupDNAs {

		groupArticle := &models.GroupArticle{}
		groupArticle.GroupDNA = groupDNA
		groupArticle.ArticleDNA = shareBatch.ArticleDNA
		groupArticle.MemberAddress = shareBatch.GroupMemberAddress

		db.Where(groupArticle).First(&groupArticle)

		if groupArticle.ID == 0 {
			groupArticle.CreatedAt = uint(time.Now().Unix())
		}

		groupArticle.TxStatus = models.TxStatusConfirmed

		db.Save(&groupArticle)

		group := &models.Group{ DNA: groupDNA }

		db.Set("gorm:query_option", "FOR UPDATE").Where(group).First(group)

		group.ArticleCount = group.ArticleCount + 1

		db.Save(group)

		models.ShareArticleIncentive(groupArticle, db)
	}

	article := &models.Article{}
	db.Set("gorm:query_option", "FOR UPDATE").Where(&models.Article{DNA: shareBatch.ArticleDNA}).First(article)

	if article.ID == 0 {
		return errors.New("article does not exist")
	}

	article.ShareCount = article.ShareCount + uint(len(shareBatch.GroupDNAs))

	db.Set("gorm:save_associations", false).Save(article)

	return nil
}
