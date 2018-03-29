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

package models

import (
	"time"
	"github.com/jinzhu/gorm"
	"math/big"
	"github.com/shopspring/decimal"
	"log"
)

const IncentiveFromArticle = 1
const IncentiveFromGroup = 2
const IncentiveFromLike  = 3
const IncentiveFromComment = 4
const IncentiveFromShare = 5

const IncentivesPending = 1
const IncentivesCalculating = 2
const IncentivesPaid = 3

type Incentive struct {
	ID              uint `gorm:"primary_key"`
	CreatedAt       uint
	IncentiveType   uint `gorm:"index"`
	UserAddress     string `gorm:"index"`
	ArticleDNA      string `gorm:"index"`
	GroupDNA        string `gorm:"index"`
	Amount          decimal.Decimal `gorm:"type:decimal(65)"`
	Status          uint `gorm:"index"`
	Score           decimal.Decimal `gorm:"type:decimal(65)"`

	// Relations
	IncentiveArticle Article   `gorm:"ForeignKey:ArticleDNA;AssociationForeignKey:DNA"`
	IncentiveGroup   Group     `gorm:"ForeignKey:GroupDNA;AssociationForeignKey:DNA"`
}

type GroupIncentive struct {
	ID              uint `gorm:"primary_key"`
	CreatedAt       uint
	GroupDNA        string `gorm:"index"`
	Amount          decimal.Decimal `gorm:"type:decimal(65)"`
	Status          uint `gorm:"index"`
	AvgScore        decimal.Decimal `gorm:"type:decimal(65)"`
	AvgCount        uint
}

type TotalScore struct {
	Score decimal.Decimal `gorm:"type:decimal(65)"`
}

func LikeArticleIncentive(like *ArticleLike, db *gorm.DB) {

	u := &User{ Address: like.GroupMemberAddress }
	db.Where(u).First(&u)
	hp := u.GetHP(db)

	inc := newIncentive()
	inc.IncentiveType = IncentiveFromLike
	inc.UserAddress = like.GroupMemberAddress
	inc.ArticleDNA = like.ArticleDNA
	inc.GroupDNA = like.GroupDNA

	inc.Score = decimal.NewFromBigInt(hp, 0)

	db.Set("gorm:save_associations", false).Save(inc)

	likeWeight := big.NewInt(1)

	// Update article score
	updateArticleScore(like.ArticleDNA, hp.Mul(hp, likeWeight), db)

	// Update group score
}

func CommentArticleIncentive(comment *ArticleComment, db *gorm.DB) {

	u := &User{ Address: comment.GroupMemberAddress }
	db.Where(u).First(&u)
	hp := u.GetHP(db)

	inc := newIncentive()
	inc.IncentiveType = IncentiveFromComment
	inc.UserAddress = comment.GroupMemberAddress
	inc.Score = decimal.NewFromBigInt(hp, 0)
	inc.ArticleDNA = comment.ArticleDNA
	inc.GroupDNA = comment.GroupDNA

	db.Set("gorm:save_associations", false).Save(inc)

	commentWeight := big.NewInt(10)

	// Update article score
	updateArticleScore(comment.ArticleDNA, hp.Mul(hp, commentWeight), db)
}

func ShareArticleIncentive(share *GroupArticle, db *gorm.DB) {

	u := &User{ Address: share.MemberAddress }
	db.Where(u).First(&u)

	if u.ID == 0 {
		log.Println("user does not exist: " + share.MemberAddress)
	}

	hp := u.GetHP(db)

	inc := newIncentive()
	inc.IncentiveType = IncentiveFromShare
	inc.UserAddress = share.MemberAddress
	inc.Score = decimal.NewFromBigInt(hp, 0)
	inc.ArticleDNA = share.ArticleDNA
	inc.GroupDNA = share.GroupDNA

	db.Set("gorm:save_associations", false).Save(inc)

	shareWeight := big.NewInt(100)

	// Update article score
	updateArticleScore(share.ArticleDNA, hp.Mul(hp, shareWeight), db)
}

func newIncentive() *Incentive {
	inc := &Incentive{}
	inc.CreatedAt = uint(time.Now().Unix())
	inc.Amount = decimal.Zero
	inc.Score = decimal.Zero
	inc.Status = IncentivesPending

	return inc
}

func updateArticleScore(articleDNA string, increment *big.Int, db *gorm.DB) {
	articleInc := &Incentive{}

	in := db.Set("gorm:query_option", "FOR UPDATE").Table("incentives")
	in = in.Where("article_dna = ?", articleDNA).Where("incentives.incentive_type = ?", IncentiveFromArticle)
	in = in.Where("incentives.status = ?", IncentivesPending)
	in.Find(articleInc)

	if articleInc.ID == 0 {

		// New incentives for today

		article := &Article{DNA: articleDNA}

		db.Where(article).First(article)

		articleInc.IncentiveType = IncentiveFromArticle
		articleInc.ArticleDNA = article.DNA
		articleInc.UserAddress = article.UserAddress
		articleInc.Status = IncentivesPending
		articleInc.CreatedAt = uint(time.Now().Unix())
		articleInc.Amount = decimal.Zero
		articleInc.Score = decimal.Zero
		articleInc.Status = IncentivesPending
	}

	incrementDecimal := decimal.NewFromBigInt(increment, 0)

	articleInc.Score = articleInc.Score.Add(incrementDecimal)

	db.Set("gorm:save_associations", false).Save(articleInc)
}
