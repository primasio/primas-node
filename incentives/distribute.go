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

package incentives

import (
	"github.com/primasio/primas-node/models"
	"github.com/jinzhu/gorm"
	"math"
	"math/big"
	"github.com/shopspring/decimal"
)

func DistributeIncentives(totalIncentivesToday *big.Int, db *gorm.DB) {

	mul4 := totalIncentivesToday.Mul(big.NewInt(4), totalIncentivesToday)

	// Lock current pending incentive records

	in := db.Table("incentives").Where("status = ?", models.IncentivesPending)
	in.Updates(map[string]interface{}{"status": models.IncentivesCalculating})

	percent40 := mul4.Div(mul4, big.NewInt(10))

	// Calculate incentive values

	// Fix amount of distribution for now

	fixedAmount := new(big.Int)
	fixedAmount.SetString("10800000000000000000000", 10)

	calculateArticleIncentivesForToday(fixedAmount, db)                                   // 40% for articles
	calculateGroupIncentivesForToday(percent40, db)                                     // 40% for groups
	calculateNodeIncentivesForToday(percent40.Div(percent40, big.NewInt(2)), db)     // 20% for nodes
}

func calculateArticleIncentivesForToday(totalIncentivesAmount *big.Int, db *gorm.DB) {

	// Align article score according to Zipf's law

	batchSize := 200
	currentBatchOffset := 0
	currentRank := 0
	currentRankScore := big.NewInt(0)
	totalScore := big.NewInt(0)

	for {
		var incentives []models.Incentive

		in := db.Table("incentives").Where("status = ?", models.IncentivesCalculating)
		in = in.Where("incentive_type = ?", models.IncentiveFromArticle)
		in = in.Where("score <> 0")
		in = in.Order("score desc").Offset(currentBatchOffset).Limit(batchSize)
		in.Find(&incentives)

		if len(incentives) == 0 {
			break
		}

		for _, incentive := range incentives {

			incentiveScore := incentive.Score.Coefficient()

			if currentRankScore.Cmp(big.NewInt(0)) == 0 || incentiveScore.Cmp(currentRankScore) < 0 {
				currentRank = currentRank + 1
				currentRankScore = incentiveScore
			}

			zipfCoefficient := big.NewInt(int64(math.Ceil(math.Sqrt(float64(currentRank)))))

			incentiveScore.Div(incentiveScore, zipfCoefficient)

			if incentiveScore.Cmp(big.NewInt(0)) == 0 {
				incentive.Score = decimal.NewFromBigInt(big.NewInt(1), 0)
			} else {
				incentive.Score = decimal.NewFromBigInt(incentiveScore, 0)
			}

			db.Save(incentive)

			totalScore = totalScore.Add(totalScore, incentiveScore)
		}

		currentBatchOffset = currentBatchOffset + batchSize
	}

	// Calculate incentives
	currentBatchOffset = 0

	for {
		var incentives []models.Incentive

		in := db.Table("incentives").Where("status = ?", models.IncentivesCalculating)
		in = in.Where("incentive_type = ?", models.IncentiveFromArticle)
		in = in.Where("score <> 0")
		in = in.Order("score desc").Offset(currentBatchOffset).Limit(batchSize)
		in.Find(&incentives)

		if len(incentives) == 0 {
			break
		}

		for _, incentive := range incentives {

			amount := big.NewInt(1)

			incScore := incentive.Score.Coefficient()

			amount = amount.Mul(totalIncentivesAmount, incScore)
			amount = amount.Div(amount, totalScore)

			amountDecimal := decimal.NewFromBigInt(amount, 0)

			amountDecimalToContributors := amountDecimal.Div(decimal.NewFromBigInt(big.NewInt(10), 0)).Floor()

			incentive.Amount = amountDecimal.Sub(amountDecimalToContributors)

			db.Save(incentive)

			article := &models.Article{ DNA: incentive.ArticleDNA }

			db.Set("gorm:query_option", "FOR UPDATE").Where(article).First(article)
			article.TotalIncentives = article.TotalIncentives.Add(incentive.Amount)
			db.Save(article)

			// Calculate article contributor incentives
			calculateArticleContributorIncentives(&incentive, amountDecimalToContributors.Coefficient(), db)
		}

		currentBatchOffset = currentBatchOffset + batchSize
	}
}

func calculateArticleContributorIncentives(articleIncentive *models.Incentive, totalAmount *big.Int, db *gorm.DB) {

	totalScore := &models.TotalScore{}

	in := db.Table("incentives").Where("status = ?", models.IncentivesCalculating)
	in = in.Where("incentive_type in (?)",[]int{models.IncentiveFromLike, models.IncentiveFromComment, models.IncentiveFromShare})
	in = in.Where("article_dna = ?", articleIncentive.ArticleDNA)
	in = in.Select("SUM(score) as score")

	in.Scan(totalScore)

	batchSize := 200
	currentBatchOffset := 0
	totalScoreInt := totalScore.Score.Coefficient()

	for {
		var incentives []models.Incentive

		in := db.Table("incentives").Where("status = ?", models.IncentivesCalculating)
		in = in.Where("incentive_type in (?)",[]int{models.IncentiveFromLike, models.IncentiveFromComment, models.IncentiveFromShare})
		in = in.Where("article_dna = ?", articleIncentive.ArticleDNA)
		in = in.Order("created_at desc").Offset(currentBatchOffset).Limit(batchSize)
		in.Find(&incentives)

		if len(incentives) == 0 {
			break
		}

		for _, incentive := range incentives {
			amount := big.NewInt(1)

			incScore := incentive.Score.Coefficient()

			amount = amount.Mul(totalAmount, incScore)

			amount = amount.Div(amount, totalScoreInt)

			incentive.Amount = decimal.NewFromBigInt(amount, 0)

			db.Save(incentive)
		}

		currentBatchOffset = currentBatchOffset + batchSize
	}
}

func calculateGroupIncentivesForToday(totalIncentivesAmount *big.Int, db *gorm.DB) {

}

func calculateNodeIncentivesForToday(totalIncentivesAmount *big.Int, db *gorm.DB) {

}
