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

package incentives_test

import (
	"testing"
	"github.com/primasio/primas-node/incentives"
	"github.com/primasio/primas-node/tests"
	"github.com/magiconair/properties/assert"
	"github.com/primasio/primas-node/db"
	"github.com/primasio/primas-node/models"
	"github.com/primasio/primas-node/cron"
	"math/big"
	"github.com/shopspring/decimal"
)

func TestArticleScoreCalculation (t *testing.T) {
	tests.InitTestEnv("../config/")

	owner, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	article1, err := tests.CreateTestArticle(owner)
	assert.Equal(t, err, nil)

	article2, err := tests.CreateTestArticle(owner)
	assert.Equal(t, err, nil)

	article3, err := tests.CreateTestArticle(owner)
	assert.Equal(t, err, nil)

	article4, err := tests.CreateTestArticle(owner)
	assert.Equal(t, err, nil)

	article5, err := tests.CreateTestArticle(owner)
	assert.Equal(t, err, nil)


	group, err := tests.CreateTestGroup(owner)
	assert.Equal(t, err, nil)


	user1, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	user1.Balance, err = decimal.NewFromString("100000000000000000000")
	assert.Equal(t, err, nil)

	user2, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)
	user2.Balance, err = decimal.NewFromString("4000000000000000000000")

	user3, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)
	user3.Balance, err = decimal.NewFromString("300000000000000000000")

	user4, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)
	user4.Balance, err = decimal.NewFromString("2100000000000000000000")

	user5, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)
	user5.Balance, err = decimal.NewFromString("1000000000000000000000")


	groupMember1, err := tests.CreateGroupMember(user1, group)
	assert.Equal(t, err, nil)

	groupMember2, err := tests.CreateGroupMember(user1, group)
	assert.Equal(t, err, nil)

	groupMember3, err := tests.CreateGroupMember(user1, group)
	assert.Equal(t, err, nil)

	groupMember4, err := tests.CreateGroupMember(user1, group)
	assert.Equal(t, err, nil)

	groupMember5, err := tests.CreateGroupMember(user1, group)
	assert.Equal(t, err, nil)


	groupArticle1, err := tests.CreateGroupArticle(article1, group, user3)
	assert.Equal(t, err, nil)

	groupArticle2, err := tests.CreateGroupArticle(article2, group, user3)
	assert.Equal(t, err, nil)

	groupArticle3, err := tests.CreateGroupArticle(article3, group, user3)
	assert.Equal(t, err, nil)

	groupArticle4, err := tests.CreateGroupArticle(article4, group, user3)
	assert.Equal(t, err, nil)

	groupArticle5, err := tests.CreateGroupArticle(article5, group, user3)
	assert.Equal(t, err, nil)


	dbi := db.GetDb()

	dbi.Save(owner)
	dbi.Save(group)

	dbi.Save(user1)
	dbi.Save(user2)
	dbi.Save(user3)
	dbi.Save(user4)
	dbi.Save(user5)

	dbi.Save(article1)
	dbi.Save(article2)
	dbi.Save(article3)
	dbi.Save(article4)
	dbi.Save(article5)

	dbi.Save(groupMember1)
	dbi.Save(groupMember2)
	dbi.Save(groupMember3)
	dbi.Save(groupMember4)
	dbi.Save(groupMember5)

	dbi.Save(groupArticle1)
	dbi.Save(groupArticle2)
	dbi.Save(groupArticle3)
	dbi.Save(groupArticle4)
	dbi.Save(groupArticle5)


	//likeA1U1, err := tests.CreateArticleLike(article1, group, user1)
	//assert.Equal(t, err, nil)
	//
	//likeA2U1, err := tests.CreateArticleLike(article2, group, user1)
	//assert.Equal(t, err, nil)
	//
	//likeA3U1, err := tests.CreateArticleLike(article3, group, user1)
	//assert.Equal(t, err, nil)
	//
	//likeA4U1, err := tests.CreateArticleLike(article4, group, user1)
	//assert.Equal(t, err, nil)
	//
	//likeA5U1, err := tests.CreateArticleLike(article5, group, user1)
	//assert.Equal(t, err, nil)
	//
	//likeA1U2, err := tests.CreateArticleLike(article1, group, user2)
	//assert.Equal(t, err, nil)
	//
	//likeA2U2, err := tests.CreateArticleLike(article2, group, user2)
	//assert.Equal(t, err, nil)
	//
	//likeA3U2, err := tests.CreateArticleLike(article3, group, user2)
	//assert.Equal(t, err, nil)
	//
	//
	//dbi.Save(likeA1U1)
	//dbi.Save(likeA2U1)
	//dbi.Save(likeA3U1)
	//dbi.Save(likeA4U1)
	//dbi.Save(likeA5U1)
	//dbi.Save(likeA1U2)
	//dbi.Save(likeA2U2)
	//dbi.Save(likeA3U2)


	// Add some article incentives

	models.ShareArticleIncentive(groupArticle1, dbi)
	models.ShareArticleIncentive(groupArticle2, dbi)
	models.ShareArticleIncentive(groupArticle3, dbi)
	models.ShareArticleIncentive(groupArticle4, dbi)
	models.ShareArticleIncentive(groupArticle5, dbi)

	//models.LikeArticleIncentive(likeA1U1, dbi)
	//models.LikeArticleIncentive(likeA2U1, dbi)
	//models.LikeArticleIncentive(likeA3U1, dbi)
	//models.LikeArticleIncentive(likeA4U1, dbi)
	//models.LikeArticleIncentive(likeA5U1, dbi)
	//models.LikeArticleIncentive(likeA1U2, dbi)
	//models.LikeArticleIncentive(likeA2U2, dbi)
	//models.LikeArticleIncentive(likeA3U2, dbi)
}

func TestArticleIncentivesDistribution (t *testing.T) {

	tests.InitTestEnv("../config/")

	totalIncentives := big.NewInt(0)
	totalIncentives.SetString("200000000000000000000000", 10)

	dbi := db.GetDb()

	incentives.DistributeIncentives(totalIncentives, dbi)
}

func TestInflate (t *testing.T) {

	tests.InitTestEnv("../config/")

	cron.TriggerInflation()
}