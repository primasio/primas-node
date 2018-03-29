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
	"strconv"
)

type IncentiveController struct{}

func (incentiveCtrl *IncentiveController) List (c *gin.Context) {

	addr := c.Query("address")

	if addr == "" {
		Error("invalid parameters", c)
	}

	offsetNum := 0
	offset := c.Query("offset")

	if offset != "" {
		if num, err := strconv.Atoi(offset); err == nil {
			offsetNum = num
		}
	}

	var incentives [] models.Incentive

	in := db.GetDb().Where(&models.Incentive{UserAddress: addr, Status:models.IncentivesPaid})
	in = in.Preload("IncentiveGroup").Preload("IncentiveArticle")
	in = in.Order("created_at desc")
	in = in.Offset(offsetNum).Limit(20)
	in.Find(&incentives)

	Success(incentives, c)
}

func (incentiveCtrl *IncentiveController) GetUserTotalIncentive (c *gin.Context) {
	address := c.Param("address")

	if address == "" {
		Error("invalid parameters", c)
	}

	from := c.Query("from")
	to := c.Query("to")

	incentive := &models.Incentive{ UserAddress: address, Status:models.IncentivesPaid }

	in := db.GetDb().Where(incentive)
	in = in.Where("created_at >= ?", from)
	in = in.Where("created_at < ?", to)
	in = in.Select("SUM(amount) AS score")

	totalScore := &models.TotalScore{}

	in.Scan(totalScore)

	Success(totalScore.Score, c)
}
