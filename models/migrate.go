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

import "github.com/primasio/primas-node/db"

func MigrateAll () {

	instance := db.GetDb()

	instance.AutoMigrate(&Article{})
	instance.AutoMigrate(&ArticleContent{})
	instance.AutoMigrate(&ArticleLike{})
	instance.AutoMigrate(&ArticleComment{})
	instance.AutoMigrate(&User{})
	instance.AutoMigrate(&System{})
	instance.AutoMigrate(&Group{})
	instance.AutoMigrate(&GroupMember{})
	instance.AutoMigrate(&GroupArticle{})
	instance.AutoMigrate(&TokenLock{})
	instance.AutoMigrate(&Incentive{})
	instance.AutoMigrate(&GroupIncentive{})
}