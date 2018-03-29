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

package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/primasio/primas-node/models"
	"github.com/primasio/primas-node/db"
	"time"
	"github.com/primasio/primas-node/crypto"
	"strconv"
	"github.com/primasio/primas-node/contracts"
)

type ArticleController struct{}

func (ctrl ArticleController) Publish(c *gin.Context) {

	var article models.Article

	if err := c.Bind(&article); err == nil {

		dbInstance := db.GetDb().Begin()

		// Generate content hash

		if article.ContentHash, err = article.GenerateContentHash(); err != nil {
			dbInstance.Rollback()
			Error(err.Error(), c)
			return
		}

		// Validate author signature

		author, err := crypto.ExtractUserFromSignature(article.GetSignatureBaseString(), article.Signature)

		if err != nil {
			dbInstance.Rollback()
			Error(err.Error(), c)
			return
		}

		if author.Address != article.UserAddress {
			dbInstance.Rollback()
			ErrorSignature(c)
			return
		}

		models.IdentifyUser(author, dbInstance)

		// Write current block hash
		latestBlockHash := models.GetState("CurrentBlockHash", dbInstance)

		if latestBlockHash == "" {
			dbInstance.Rollback()
			Error("node not synchronized yet", c)
			return
		}

		article.BlockHash = latestBlockHash

		// Generate article DNA
		if article.DNA, err = article.GenerateDNA(); err != nil {
			dbInstance.Rollback()
			Error(err.Error(), c)
			return
		}

		duplicateCheck := &models.Article { DNA: article.DNA }

		dbInstance.Where(duplicateCheck).First(duplicateCheck)

		if duplicateCheck.ID != 0 {
			dbInstance.Rollback()
			Error("article " + article.DNA + " already exists", c)
			return
		}

		// Generate article content
		articleContent := models.NewArticleContent(article.DNA, article.Content)
		dbInstance.Save(&articleContent)

		// Generate article extra data
		if article.Abstract, err = article.GenerateAbstract(); err != nil {
			dbInstance.Rollback()
			Error(err.Error(), c)
			return
		}

		article.CreatedAt = uint(time.Now().Unix())
		article.TxStatus = models.TxStatusPending

		// Save metadata on Blockchain

		contentContract, err2 := contracts.GetContentContract()

		if err2 != nil {
			dbInstance.Rollback()
			Error(err2.Error(), c)
			return
		}

		err3 := contentContract.Publish(&article)

		if err3 != nil {
			dbInstance.Rollback()
			Error(err3.Error(), c)
			return
		}

		// Save article in database
		dbInstance.Set("gorm:save_associations", false).Save(&article)

		dbInstance.Commit()
		Success(article, c)

	} else {
		Error(err.Error(), c)
	}
}

func (ctrl *ArticleController) Get(c *gin.Context) {
	dna := c.Param("dna")

	article := &models.Article{ DNA: dna }

	db.GetDb().Preload("Author").Where(article).First(&article)

	if article.ID == 0 {
		ErrorNotFound("article does not exist", c)
		return
	}

	Success(article, c)
}

func (ctrl *ArticleController) GetContent(c *gin.Context) {
	dna := c.Param("dna")

	articleContent := &models.ArticleContent{ DNA: dna }

	db.GetDb().Where(articleContent).First(&articleContent)

	if articleContent.ID == 0 {
		ErrorNotFound("article does not exist", c)
		return
	}

	Success(articleContent.Content, c)
}

func (ctrl *ArticleController) List(c *gin.Context) {

	address := c.Query("address")
	start := c.Query("start")

	if address == "" || start == "" {
		Error("invalid parameters", c)
		return
	}

	offsetNum := 0
	offset := c.Query("offset")

	if offset != "" {
		if num, err := strconv.Atoi(offset); err == nil {
			offsetNum = num
		}
	}

	var articles []models.Article

	dbi := db.GetDb()

	in := dbi.Table("articles").Preload("Author")
	in = in.Joins("join group_articles on group_articles.article_dna=articles.dna")
	in = in.Joins("join group_members on group_members.group_dna=group_articles.group_dna")
	in = in.Where("group_members.member_address = ?", address)
	in = in.Where("articles.created_at <= ?", start)
	in = in.Order("group_articles.created_at desc")
	in = in.Offset(offsetNum)
	in = in.Limit(20)
	in = in.Select("articles.*, group_articles.group_dna")
	in.Find(&articles)

	Success(articles, c)
}

func (ctrl *ArticleController) Discover(c *gin.Context) {
	var articles []models.Article

	dbi := db.GetDb()

	in := dbi.Table("articles").Preload("Author")
	in = in.Joins("join group_articles on group_articles.article_dna=articles.dna")
	in = in.Select("articles.*, group_articles.group_dna")
	in = in.Where("articles.tx_status = ?", models.TxStatusConfirmed)
	in = in.Order("RAND()")
	in = in.Limit(20)
	in.Find(&articles)

	Success(articles, c)
}