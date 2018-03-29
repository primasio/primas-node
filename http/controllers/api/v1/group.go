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
	"github.com/jinzhu/gorm"
	"time"
	"github.com/primasio/primas-node/crypto"
	"strconv"
	"github.com/primasio/primas-node/contracts"
)

type GroupController struct {}

func (groupCtrl *GroupController) Create (c *gin.Context) {

	group := models.NewGroup("", "")

	if err := c.Bind(group); err == nil {

		tx := db.GetDb().Begin()

		// Validate creator signature

		creator, err := crypto.ExtractUserFromSignature(group.GetSignatureBaseString(), group.Signature)

		if err != nil {
			tx.Rollback()
			Error(err.Error(), c)
			return
		}

		if creator.Address != group.UserAddress {
			tx.Rollback()
			ErrorSignature(c)
			return
		}

		models.IdentifyUser(creator, tx)

		// Generate group DNA
		if group.DNA, err = group.GenerateDNA(); err != nil {
			tx.Rollback()
			Error(err.Error(), c)
			return
		}

		duplicateCheck := &models.Group { DNA: group.DNA }

		tx.Where(duplicateCheck).First(duplicateCheck)

		if duplicateCheck.ID != 0 {
			tx.Rollback()
			Error("group " + group.DNA + " already exists", c)
			return
		}

		// Add creator to group member

		groupMember := &models.GroupMember{}
		groupMember.GroupDNA = group.DNA
		groupMember.MemberAddress = creator.Address
		groupMember.TxStatus = models.TxStatusPending
		groupMember.CreatedAt = uint(time.Now().Unix())

		tx.Save(&groupMember)

		groupContract, err := contracts.GetGroupContract()

		if err != nil {
			Error(err.Error(), c)
			return
		}

		if err := groupContract.Create(group); err != nil {
			Error(err.Error(), c)
			return
		}

		// Save group in database
		tx.Save(&group)

		tx.Commit()
		Success(group, c)

	} else {
		Error(err.Error(), c)
	}
}

func (groupCtrl *GroupController) AddMember(c *gin.Context) {

	dbi := db.GetDb()

	group := groupCtrl.retrieveGroup(c, dbi)

	if group == nil {
		ErrorNotFound("group does not exist", c)
		return
	}

	groupMember := &models.GroupMember{}

	err := c.Bind(groupMember)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	groupMember.GroupDNA = group.DNA

	// Validate Signature
	check, err := crypto.ExtractUserFromSignature(groupMember.GetSignatureBaseString(), groupMember.Signature)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	if check.Address != groupMember.MemberAddress {
		ErrorSignature(c)
		return
	}

	dbi.Where(groupMember).First(&groupMember)

	if groupMember.ID != 0 {
		Error("member already in group", c)
		return
	}

	groupMember.CreatedAt = uint(time.Now().Unix())

	groupMember.TxStatus = models.TxStatusPending

	groupContract, err := contracts.GetGroupContract()

	if err != nil {
		Error(err.Error(), c)
		return
	}

	if err := groupContract.AddMember(groupMember); err != nil {
		Error(err.Error(), c)
		return
	}

	dbi.Save(&groupMember)

	Success(groupMember, c)
}

func (groupCtrl *GroupController) RemoveMember(c *gin.Context) {

	dbi := db.GetDb()

	group := groupCtrl.retrieveGroup(c, dbi)

	if group == nil {
		ErrorNotFound("group does not exist", c)
		return
	}

	groupMember := &models.GroupMember{}

	err := c.Bind(groupMember)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	groupMember.GroupDNA = group.DNA

	check, err := crypto.ExtractUserFromSignature(groupMember.GetSignatureBaseString(), groupMember.Signature)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	if check.Address != groupMember.MemberAddress {
		ErrorSignature(c)
		return
	}

	dbi.Where(groupMember).First(&groupMember)

	if groupMember.ID == 0 {
		ErrorNotFound("member does not exist", c)
		return
	}

	if groupMember.TxStatus != models.TxStatusConfirmed {

		// Last modification is not confirmed yet
		ErrorPendingTx(c)
		return
	}

	groupContract, err := contracts.GetGroupContract()

	if err != nil {
		Error(err.Error(), c)
		return
	}

	if err := groupContract.RemoveMember(groupMember); err != nil {
		Error(err.Error(), c)
		return
	}

	// Delete should be postponed till transaction confirmed

	groupMember.TxStatus = models.TxStatusPending
	dbi.Save(&groupMember)

	Success(groupMember, c)
}

func (groupCtrl *GroupController) RemoveMemberByOwner(c *gin.Context) {

	dbi := db.GetDb()

	group := groupCtrl.retrieveGroup(c, dbi)

	if group == nil {
		ErrorNotFound("group does not exist", c)
		return
	}

	groupMember := &models.GroupMember{}

	err := c.Bind(groupMember)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	groupMember.GroupDNA = group.DNA

	check, err := crypto.ExtractUserFromSignature(groupMember.GetOwnerSignatureBaseString(), groupMember.Signature)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	if check.Address != group.UserAddress {
		ErrorSignature(c)
		return
	}

	dbi.Where(groupMember).First(&groupMember)

	if groupMember.ID == 0 {
		ErrorNotFound("member does not exist", c)
		return
	}

	if groupMember.TxStatus != models.TxStatusConfirmed {

		// Last modification is not confirmed yet
		ErrorPendingTx(c)
		return
	}

	groupContract, err := contracts.GetGroupContract()

	if err != nil {
		Error(err.Error(), c)
		return
	}

	if err := groupContract.RemoveMemberByOwner(groupMember, group.UserAddress); err != nil {
		Error(err.Error(), c)
		return
	}

	// Delete should be postponed till transaction confirmed
	groupMember.TxStatus = models.TxStatusPending
	dbi.Save(&groupMember)

	Success(groupMember, c)
}

func (groupCtrl *GroupController) Get (c *gin.Context) {

	dbi := db.GetDb()

	group := groupCtrl.retrieveGroup(c, dbi)

	if group == nil {
		ErrorNotFound("group does not exist", c)
		return
	}

	address := c.Query("address")

	if address != "" {
		groupMember := &models.GroupMember{ GroupDNA: group.DNA, MemberAddress: address }

		dbi.Where(groupMember).First(groupMember)

		if groupMember.ID != 0 {
			group.IsMember = groupMember
		}
	}

	// Get first page of articles

	var articles []models.Article

	in := dbi.Table("articles")
	in = in.Joins("join group_articles on group_articles.article_dna=articles.dna")
	in = in.Where(&models.GroupArticle{GroupDNA:group.DNA})
	in = in.Order("articles.created_at desc")
	in = in.Limit(20)

	in.Find(&articles)

	group.GroupArticles = articles

	Success(group, c)
}

func (groupCtrl *GroupController) GetArticles(c *gin.Context) {

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

	group := groupCtrl.retrieveGroup(c, dbi)

	if group == nil {
		ErrorNotFound("group does not exist", c)
		return
	}

	var articles []models.Article

	in := dbi.Table("articles")
	in = in.Joins("join group_articles on group_articles.article_dna=articles.dna")
	in = in.Where(&models.GroupArticle{GroupDNA:group.DNA})
	in = in.Where("articles.created_at <= ?", start)
	in = in.Order("created_at DESC").Offset(offsetNum).Limit(20).Find(&articles)

	Success(articles, c)
}

func (groupCtrl *GroupController) GetMembers (c *gin.Context) {
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

	group := groupCtrl.retrieveGroup(c, dbi)

	if group == nil {
		ErrorNotFound("group does not exist", c)
		return
	}

	var users []models.User

	in := dbi.Table("users")
	in = in.Joins("join group_members on group_members.member_address=users.address")
	in = in.Where(&models.GroupMember{GroupDNA: group.DNA})
	in = in.Where("group_members.created_at <= ?", start)
	in = in.Order("created_at DESC").Offset(offsetNum).Limit(20).Find(&users)

	Success(users, c)
}

func (groupCtrl *GroupController) List (c *gin.Context) {

	address := c.Query("address")

	if address == "" {
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

	var groups [] models.Group

	in := db.GetDb().Table("groups")

	in = in.Joins("join group_members on group_members.group_dna=groups.dna")
	in = in.Where(&models.GroupMember{MemberAddress:address})
	in = in.Offset(offsetNum).Limit(20).Order("created_at desc").Find(&groups)
	Success(groups, c)
}

func (groupCtrl *GroupController) Discover(c *gin.Context) {
	var groups []models.Group

	dbi := db.GetDb()

	in := dbi.Table("groups")
	in = in.Order("RAND()")
	in = in.Where("tx_status = ?", models.TxStatusConfirmed)
	in = in.Limit(20)

	in.Find(&groups)

	Success(groups, c)
}

func (groupCtrl *GroupController) retrieveGroup(c *gin.Context, db *gorm.DB) *models.Group {
	dna := c.Param("dna")

	if dna == "" {
		return nil
	}

	group := &models.Group{ DNA: dna }

	db.Where(group).First(&group)

	if group.ID == 0 || group.TxStatus == models.TxStatusPending {
		return nil
	}

	return group
}