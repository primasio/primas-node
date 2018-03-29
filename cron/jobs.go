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

package cron

import (
	"github.com/jasonlvhit/gocron"
	"github.com/primasio/primas-node/contracts"
	"log"
)

func StartCronJobs () {

	// Calculate and distribute incentives every day
	gocron.Every(1).Day().At("20:00").Do(TriggerInflation)

	<- gocron.Start()
}

func TriggerInflation() {

	tokenContract, err := contracts.GetTokenContract()

	if err != nil {
		log.Println(err.Error())
		return
	}

	if err := tokenContract.Inflate(); err != nil {
		log.Println(err.Error())
	}
}