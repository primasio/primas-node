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
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/crypto"
	"time"
	"encoding/hex"
	"strconv"
	"github.com/primasio/primas-node/contracts"
)

type UserController struct {}

func (userCtrl *UserController) Update (c *gin.Context) {
	var user models.User

	err := c.Bind(&user)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	// Validate author signature

	msg :=  crypto.Keccak256([]byte(user.Name + user.Extra))

	sigBytes, err := hex.DecodeString(user.Signature)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	publicKey, err := secp256k1.RecoverPubkey(msg, sigBytes)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	pk := crypto.ToECDSAPub(publicKey)
	addr := crypto.PubkeyToAddress(*pk)

	if addr.Hex() != user.Address {
		Error("invalid signature", c)
		return
	}

	// TODO: Update user data on Blockchain

	dbi := db.GetDb()

	check := &models.User{}
	check.Address = addr.Hex()
	dbi.Where(check).First(&check)

	if check.ID != 0 {
		// Modify existing data
		check.Name = user.Name
		check.Signature = user.Signature
		check.Extra = user.Extra
		dbi.Save(check)

		Success(check, c)
	} else {
		// Create new data
		user.Address = addr.Hex()
		user.CreatedAt = uint(time.Now().Unix())
		dbi.Save(&user)

		Success(user, c)
	}
}

func (userCtrl *UserController) Get (c *gin.Context) {
	addr := c.Param("address")

	if addr == "" {
		Error("invalid parameters", c)
		return
	}

	user := &models.User{ Address:addr }

	dbi := db.GetDb()

	dbi.Where(user).First(&user)

	if user.ID == 0 {
		ErrorNotFound("user does not exist", c)
		return
	}

	// Get user's articles
	var articles []models.Article

	in := dbi.Table("articles")
	in = in.Where(&models.Article{UserAddress: addr})
	in = in.Order("created_at desc").Limit(20)

	in.Find(&articles)

	user.UserArticles = articles

	// Get user's groups

	var groups []models.Group

	in2 := dbi.Table("groups")
	in2 = in2.Where(&models.Group{UserAddress: addr})
	in2 = in2.Order("created_at desc").Limit(20)

	in2.Find(&groups)

	user.UserGroups = groups

	Success(user, c)
}

func (userCtrl *UserController) GetGroups (c *gin.Context) {
	addr := c.Param("address")

	if addr == "" {
		Error("invalid parameters", c)
		return
	}

	dbi := db.GetDb()

	offsetNum := 0
	offset := c.Query("offset")

	if offset != "" {
		if num, err := strconv.Atoi(offset); err == nil {
			offsetNum = num
		}
	}

	var groups [] models.Group

	dbi.Where(&models.Group{ UserAddress: addr}).Order("created_at desc").Offset(offsetNum).Limit(20).Find(&groups)

	Success(groups, c)
}

func (userCtrl *UserController) GetArticles (c *gin.Context) {
	addr := c.Param("address")

	if addr == "" {
		Error("invalid parameters", c)
		return
	}

	dbi := db.GetDb()

	offsetNum := 0
	offset := c.Query("offset")

	if offset != "" {
		if num, err := strconv.Atoi(offset); err == nil {
			offsetNum = num
		}
	}

	var articles [] models.Article

	dbi.Where(&models.Article{ UserAddress: addr}).Order("created_at desc").Offset(offsetNum).Limit(20).Find(&articles)

	Success(articles, c)
}

func (userCtrl *UserController) GetGroupArticles (c *gin.Context) {

	addr := c.Param("address")
	dna := c.Param("dna")

	if addr == "" || dna == "" {
		Error("invalid parameters", c)
		return
	}

	dbi := db.GetDb()

	offsetNum := 0
	offset := c.Query("offset")

	if offset != "" {
		if num, err := strconv.Atoi(offset); err == nil {
			offsetNum = num
		}
	}

	var articles [] models.Article

	in := dbi.Table("articles")
	in = in.Joins("join group_articles on group_articles.article_dna = articles.dna")
	in = in.Where(&models.GroupArticle{GroupDNA: dna, MemberAddress: addr})
	in = in.Order("created_at desc").Offset(offsetNum).Limit(20)

	in.Find(&articles)

	Success(articles, c)
}

func (userCtrl *UserController) GetBalance(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		Error("invalid parameters", c)
		return
	}

	user := &models.User{ Address:addr }

	dbi := db.GetDb()

	dbi.Where(user).First(user)

	if user.ID == 0 {
		ErrorNotFound("user does not exist", c)
		return
	}

	balance := user.GetSpendableBalance(dbi)

	Success(balance.String(), c)
}

func (userCtrl *UserController) GetLockedBalance(c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		Error("invalid parameters", c)
		return
	}

	user := &models.User{ Address:addr }

	dbi := db.GetDb()

	dbi.Where(user).First(user)

	if user.ID == 0 {
		ErrorNotFound("user does not exist", c)
		return
	}

	balance := user.GetLockedBalance(dbi)

	Success(balance.String(), c)
}

func (userCtrl *UserController) Burn (c *gin.Context) {

	addr := c.Param("address")
	if addr == "" {
		Error("invalid parameters", c)
		return
	}

	timestamp := c.PostForm("Timestamp")
	signature := c.PostForm("Signature")

	if timestamp == "" || signature == "" {
		Error("invalid parameters", c)
		return
	}

	userContract, err := contracts.GetUserContract()

	if err != nil {
		Error(err.Error(), c)
		return
	}

	if err := userContract.Burn(timestamp, addr, signature); err != nil {
		Error(err.Error(), c)
		return
	}

	Success(nil, c)
}

func (userCtrl *UserController) GetHP (c *gin.Context) {
	addr := c.Param("address")
	if addr == "" {
		Error("invalid parameters", c)
		return
	}

	user := &models.User{ Address:addr }

	dbi := db.GetDb()

	dbi.Where(user).First(user)

	if user.ID == 0 {
		ErrorNotFound("user does not exist", c)
		return
	}

	hp := user.GetHP(dbi).String()
	Success(hp, c)
}
