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
	"net/http"
	"encoding/json"
)

func Error(msg string, c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": msg})
}

func Success(data interface{}, c *gin.Context) {
	bytes, err := json.Marshal(data)

	if err != nil {
		Error(err.Error(), c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": string(bytes)})
}

func ErrorNotFound(msg string, c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"success": false, "message": msg})
}

func ErrorSignature(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid signature"})
}

func ErrorPendingTx(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Last transaction is not confirmed yet. Please try again later."})
}