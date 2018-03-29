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
	"github.com/jinzhu/gorm"
	"github.com/primasio/primas-node/models"
	"github.com/primasio/primas-node/db"
	"github.com/primasio/primas-node/crypto"
	"time"
	"errors"
	"github.com/primasio/primas-node/contracts"
	"strconv"
)

type ArticleInteractController struct {}

func (articleInteractCtrl *ArticleInteractController) Like (c *gin.Context) {

	dbi := db.GetDb()

	articleLike := &models.ArticleLike{}

	err := c.Bind(articleLike)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	group, article, err := articleInteractCtrl.retrieveArticleAndGroup(articleLike.GroupDNA, c, dbi)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	articleLike.ArticleDNA = article.DNA

	// Check signature

	check, err := crypto.ExtractUserFromSignature(articleLike.GetSignatureBaseString(), articleLike.Signature)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	if check.Address != articleLike.GroupMemberAddress {
		ErrorSignature(c)
		return
	}

	checkLike := &models.ArticleLike{ ArticleDNA: articleLike.ArticleDNA, GroupDNA: articleLike.GroupDNA, GroupMemberAddress: articleLike.GroupMemberAddress }

	dbi.Where(checkLike).First(checkLike)

	if checkLike.ID != 0 {
		Error("like already clicked", c)
		return
	}

	// Check group membership
	groupMember := &models.GroupMember{GroupDNA: group.DNA, MemberAddress: check.Address}

	dbi.Where(groupMember).First(&groupMember)

	if groupMember.ID == 0 || groupMember.TxStatus == models.TxStatusPending {
		Error("user not in group", c)
		return
	}

	articleLike.TxStatus = models.TxStatusPending
	articleLike.CreatedAt = uint(time.Now().Unix())

	contentContract, err := contracts.GetContentContract()

	if err != nil {
		Error(err.Error(), c)
		return
	}

	if err := contentContract.Like(articleLike); err != nil {
		Error(err.Error(), c)
		return
	}

	// Write to db
	dbi.Save(&articleLike)

	Success(articleLike, c)
}

func (articleInteractCtrl *ArticleInteractController) Comment (c *gin.Context) {

	dbi := db.GetDb()

	articleComment := &models.ArticleComment{}

	err := c.Bind(articleComment)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	group, article, err := articleInteractCtrl.retrieveArticleAndGroup(articleComment.GroupDNA, c, dbi)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	articleComment.ArticleDNA = article.DNA

	if contentHash, err := articleComment.GenerateContentHash(); err == nil {
		articleComment.ContentHash = contentHash
	} else {
		Error(err.Error(), c)
		return
	}

	// Check signature

	check, err := crypto.ExtractUserFromSignature(articleComment.GetSignatureBaseString(), articleComment.Signature)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	if check.Address != articleComment.GroupMemberAddress {
		ErrorSignature(c)
		return
	}

	if contentHash, err := articleComment.GenerateContentHash(); err == nil {
		articleComment.ContentHash = contentHash
	} else {
		Error(err.Error(), c)
		return
	}

	// Check group membership
	groupMember := &models.GroupMember{GroupDNA: group.DNA, MemberAddress: check.Address}

	dbi.Where(groupMember).First(&groupMember)

	if groupMember.ID == 0 || groupMember.TxStatus == models.TxStatusPending {
		Error("user not in group", c)
		return
	}

	// Check duplicate content

	checkComment := &models.ArticleComment{
		ArticleDNA: articleComment.ArticleDNA,
		GroupDNA: articleComment.GroupDNA,
		GroupMemberAddress: articleComment.GroupMemberAddress,
		ContentHash: articleComment.ContentHash }

	dbi.Where(checkComment).First(checkComment)

	if checkComment.ID != 0 {
		Error("same comment already published", c)
		return
	}

	articleComment.TxStatus = models.TxStatusPending
	articleComment.CreatedAt = uint(time.Now().Unix())

	contentContract, err := contracts.GetContentContract()

	if err != nil {
		Error(err.Error(), c)
		return
	}

	if err:= contentContract.Comment(articleComment); err != nil {
		Error(err.Error(), c)
		return
	}

	dbi.Set("gorm:save_associations", false).Save(&articleComment)

	Success(articleComment, c)
}

func (articleInteractCtrl *ArticleInteractController) Share (c *gin.Context) {

	dbi := db.GetDb()

	articleShareBatch := &models.ArticleShareBatch{}

	err := c.Bind(articleShareBatch)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	article := articleInteractCtrl.retrieveArticle(c, dbi)

	if article == nil {
		Error("article does not exist", c)
		return
	}

	articleShareBatch.ArticleDNA = article.DNA

	// Check signature

	sigBase := articleShareBatch.GetSignatureBaseString()

	check, err := crypto.ExtractUserFromSignature(sigBase, articleShareBatch.Signature)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	if check.Address != articleShareBatch.GroupMemberAddress {
		ErrorSignature(c)
		return
	}

	// Check group membership

	for _, groupDNA := range articleShareBatch.GroupDNAs {

		cc := &models.GroupMember{ MemberAddress: articleShareBatch.GroupMemberAddress, GroupDNA: groupDNA }

		dbi.Where(cc).First(&cc)

		if cc.ID == 0 {
			Error("user not in group", c)
			return
		}

		ac := &models.GroupArticle{ GroupDNA: groupDNA, ArticleDNA: articleShareBatch.ArticleDNA }

		dbi.Where(ac).First(ac)

		if ac.ID != 0 {
			Error("article already shared in this group", c)
			return
		}
	}

	contentContract, err := contracts.GetContentContract()

	if err != nil {
		Error(err.Error(), c)
		return
	}

	if err := contentContract.Share(articleShareBatch); err != nil {
		Error(err.Error(), c)
		return
	}

	for _, groupDNA := range articleShareBatch.GroupDNAs {
		groupArticle := &models.GroupArticle{}
		groupArticle.GroupDNA = groupDNA
		groupArticle.MemberAddress = articleShareBatch.GroupMemberAddress
		groupArticle.ArticleDNA = article.DNA
		groupArticle.CreatedAt = uint(time.Now().Unix())
		groupArticle.TxStatus = models.TxStatusPending
		dbi.Save(groupArticle)
	}

	Success(articleShareBatch, c)
}

func (articleInteractCtrl *ArticleInteractController) GetComments(c *gin.Context) {

	start := c.Query("start")

	if start == "" {
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

	dbi := db.GetDb()

	article := articleInteractCtrl.retrieveArticle(c, dbi)

	if article == nil {
		Error("article does not exist", c)
		return
	}

	var comments []models.ArticleComment

	in := dbi.Where(&models.ArticleComment{ArticleDNA: article.DNA}).Preload("Author")
	in = in.Where("created_at <= ?", start)
	in = in.Order("created_at desc")
	in = in.Offset(offsetNum)
	in = in.Limit(20)
	in.Find(&comments)

	Success(comments, c)
}

func (articleInteractCtrl *ArticleInteractController) retrieveArticle(c *gin.Context, db *gorm.DB) *models.Article {
	dna := c.Param("dna")
	article := &models.Article{ DNA: dna }

	db.Where(article).First(&article)

	if article.ID == 0 || article.TxStatus == models.TxStatusPending {
		return nil
	}

	return article
}

func (articleInteractCtrl *ArticleInteractController) retrieveGroup(dna string, db *gorm.DB) *models.Group {
	group := &models.Group{ DNA: dna }

	db.Where(group).First(&group)

	if group.ID == 0 || group.TxStatus == models.TxStatusPending {
		return nil
	}

	return group
}

func (articleInteractCtrl *ArticleInteractController) retrieveArticleAndGroup(groupDNA string, c *gin.Context, db *gorm.DB) (*models.Group, *models.Article, error) {

	article := articleInteractCtrl.retrieveArticle(c, db)

	if article == nil {
		return nil, nil, errors.New("article does not exist")
	}

	group := articleInteractCtrl.retrieveGroup(groupDNA, db)

	if group == nil {
		return nil, nil, errors.New("group does not exist")
	}

	return group, article, nil
}