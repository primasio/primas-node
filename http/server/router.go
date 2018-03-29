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

package server

import (
	"github.com/gin-gonic/gin"
	"github.com/primasio/primas-node/http/controllers/api/v1"
)

func NewRouter() *gin.Engine {
	gin.DisableConsoleColor()

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	v1g := router.Group("v1")
	{
		userCtrl := new(v1.UserController)
		articleCtrl := new(v1.ArticleController)
		articleInteractCtrl := new(v1.ArticleInteractController)
		groupCtrl := new(v1.GroupController)
		incentiveCtrl := new(v1.IncentiveController)

		userGroup := v1g.Group("users")
		{
			userGroup.POST("", userCtrl.Update)
			userGroup.GET("/:address", userCtrl.Get)
			userGroup.GET("/:address/groups", userCtrl.GetGroups)
			userGroup.GET("/:address/groups/:dna/articles", userCtrl.GetGroupArticles)
			userGroup.GET("/:address/articles", userCtrl.GetArticles)
			userGroup.GET("/:address/balance", userCtrl.GetBalance)
			userGroup.GET("/:address/balance/locked", userCtrl.GetLockedBalance)
			userGroup.GET("/:address/hp", userCtrl.GetHP)

			userGroup.POST("/:address/burn", userCtrl.Burn)
		}

		articleGroup := v1g.Group("articles")
		{
			articleGroup.GET("", articleCtrl.List)
			articleGroup.POST("", articleCtrl.Publish)
			articleGroup.GET("/:dna/content", articleCtrl.GetContent)
			articleGroup.GET("/:dna", articleCtrl.Get)
			articleGroup.POST("/:dna/likes", articleInteractCtrl.Like)
			articleGroup.GET("/:dna/comments", articleInteractCtrl.GetComments)
			articleGroup.POST("/:dna/comments", articleInteractCtrl.Comment)
			articleGroup.POST("/:dna/groups", articleInteractCtrl.Share)
		}

		groupGroup := v1g.Group("groups")
		{
			groupGroup.GET("", groupCtrl.List)
			groupGroup.POST("", groupCtrl.Create)
			groupGroup.GET("/:dna", groupCtrl.Get)
			groupGroup.GET("/:dna/articles", groupCtrl.GetArticles)
			groupGroup.GET("/:dna/members", groupCtrl.GetMembers)
			groupGroup.POST("/:dna/members", groupCtrl.AddMember)
			groupGroup.POST("/:dna/members/delete", groupCtrl.RemoveMember)
			groupGroup.POST("/:dna/members/delete/owner", groupCtrl.RemoveMemberByOwner)
		}

		discoverGroup := v1g.Group("discover")
		{
			discoverGroup.GET("/groups", groupCtrl.Discover)
			discoverGroup.GET("/articles", articleCtrl.Discover)
		}

		incentiveGroup := v1g.Group("incentives")
		{
			incentiveGroup.GET("", incentiveCtrl.List)
			incentiveGroup.GET("/users/:address/total", incentiveCtrl.GetUserTotalIncentive)
		}
	}

	return router
}