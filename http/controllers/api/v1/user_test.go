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

package v1_test

import (
	"testing"
	"github.com/primasio/primas-node/tests"
	"github.com/magiconair/properties/assert"
	"github.com/primasio/primas-node/db"
	"github.com/ethereum/go-ethereum/crypto"
	"encoding/hex"
	"net/url"
	"strings"
	"net/http"
	"strconv"
	"github.com/primasio/primas-node/http/server"
	"net/http/httptest"
	"log"
	"time"
	"github.com/primasio/primas-node/models"
	"github.com/shopspring/decimal"
)

func TestUpdateUser(t *testing.T) {

	tests.InitTestEnv("../../../../config/")
	router := server.NewRouter()
	w := httptest.NewRecorder()

	user, pk, err := tests.CreateTestUser()

	assert.Equal(t, err, nil)

	dbi := db.GetDb()
	dbi.Save(&user)

	user.Name = "Updated"

	sigStr := crypto.Keccak256([]byte(user.Name + user.Extra))
	sigBytes, err := crypto.Sign(sigStr, pk)
	assert.Equal(t, err, nil)

	signature := hex.EncodeToString(sigBytes)

	data := url.Values{}
	data.Set("Address", user.Address)
	data.Set("Name", user.Name)
	data.Set("Extra", user.Extra)
	data.Set("Signature", signature)

	req, _ := http.NewRequest("POST", "/v1/users", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)

	reqGet, _ := http.NewRequest("GET", "/v1/users/" + user.Address, nil)

	router.ServeHTTP(w, reqGet)
	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestGetUser(t *testing.T) {
	tests.InitTestEnv("../../../../config/")
	router := server.NewRouter()
	w := httptest.NewRecorder()

	user, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	group, err := tests.CreateTestGroup(user)
	assert.Equal(t, err, nil)

	article, err := tests.CreateTestArticle(user)
	assert.Equal(t, err, nil)

	articleContent, err := tests.CreateTestArticleContent(article)
	assert.Equal(t, err, nil)

	dbi := db.GetDb()

	dbi.Save(user)
	dbi.Save(group)
	dbi.Save(article)
	dbi.Save(articleContent)

	req, _ := http.NewRequest("GET", "/v1/users/" + user.Address, nil)

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestCreateUser(t *testing.T) {
	tests.InitTestEnv("../../../../config/")
	router := server.NewRouter()
	w := httptest.NewRecorder()

	user, pk, err := tests.CreateTestUser()

	assert.Equal(t, err, nil)

	user.Name = "Created"

	sigStr := crypto.Keccak256([]byte(user.Name + user.Extra))
	sigBytes, err := crypto.Sign(sigStr, pk)
	assert.Equal(t, err, nil)

	signature := hex.EncodeToString(sigBytes)

	data := url.Values{}
	data.Set("Address", user.Address)
	data.Set("Name", user.Name)
	data.Set("Extra", user.Extra)
	data.Set("Signature", signature)

	req, _ := http.NewRequest("POST", "/v1/users", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)

	reqGet, _ := http.NewRequest("GET", "/v1/users/" + user.Address, nil)

	router.ServeHTTP(w, reqGet)
	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestGetUserGroups (t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	owner, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	group1, err := tests.CreateTestGroup(owner)
	assert.Equal(t, err, nil)

	group2, err := tests.CreateTestGroup(owner)
	assert.Equal(t, err, nil)

	dbi := db.GetDb()

	dbi.Save(owner)
	dbi.Save(group1)
	dbi.Save(group2)

	req, _ := http.NewRequest("GET", "/v1/users/" + owner.Address + "/groups", nil)

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestGetUserArticles (t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	user, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	article1, err := tests.CreateTestArticle(user)
	assert.Equal(t, err, nil)

	articleContent1, err := tests.CreateTestArticleContent(article1)
	assert.Equal(t, err, nil)

	article2, err := tests.CreateTestArticle(user)
	assert.Equal(t, err, nil)

	articleContent2, err := tests.CreateTestArticleContent(article2)
	assert.Equal(t, err, nil)


	dbi := db.GetDb()

	dbi.Save(user)
	dbi.Save(article1)
	dbi.Save(articleContent1)
	dbi.Save(article2)
	dbi.Save(articleContent2)

	req, _ := http.NewRequest("GET", "/v1/users/" + user.Address + "/articles", nil)

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestGetUserGroupArticles (t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	user, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	article1, err := tests.CreateTestArticle(user)
	assert.Equal(t, err, nil)

	articleContent1, err := tests.CreateTestArticleContent(article1)
	assert.Equal(t, err, nil)

	article2, err := tests.CreateTestArticle(user)
	assert.Equal(t, err, nil)

	articleContent2, err := tests.CreateTestArticleContent(article2)
	assert.Equal(t, err, nil)


	group, err := tests.CreateTestGroup(user)
	assert.Equal(t, err, nil)


	groupMember, err := tests.CreateGroupMember(user, group)
	assert.Equal(t, err, nil)

	groupArticle1, err := tests.CreateGroupArticle(article1, group, user)
	assert.Equal(t, err, nil)

	groupArticle2, err := tests.CreateGroupArticle(article2, group, user)
	assert.Equal(t, err, nil)

	dbi := db.GetDb()

	dbi.Save(user)
	dbi.Save(article1)
	dbi.Save(articleContent1)
	dbi.Save(article2)
	dbi.Save(articleContent2)

	dbi.Save(group)
	dbi.Save(groupMember)

	dbi.Save(groupArticle1)
	dbi.Save(groupArticle2)

	req, _ := http.NewRequest("GET", "/v1/users/" + user.Address + "/groups/" + group.DNA + "/articles", nil)

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestUserTokenBurn (t *testing.T) {

	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	keyStore, account, err := tests.LoadTestAccount(0)
	assert.Equal(t, err, nil)

	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)

	sigStr := crypto.Keccak256([]byte("burn" + timestampStr))

	sigByte, err := keyStore.SignHash(*account, []byte(sigStr))
	assert.Equal(t, err, nil)

	signature := hex.EncodeToString(sigByte)


	data := url.Values{}
	data.Set("Timestamp", timestampStr)
	data.Set("Signature", signature)

	req, _ := http.NewRequest("POST", "/v1/users/" + account.Address.Hex() + "/burn", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestUserHP (t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	user, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	article, err := tests.CreateTestArticle(user)
	assert.Equal(t, err, nil)

	articleContent, err := tests.CreateTestArticleContent(article)
	assert.Equal(t, err, nil)

	group, err := tests.CreateTestGroup(user)
	assert.Equal(t, err, nil)

	dbi := db.GetDb()

	bal, err := decimal.NewFromString("2000000000000000000000")
	assert.Equal(t, err, nil)

	user.Balance = bal

	dbi.Save(user)
	dbi.Save(article)
	dbi.Save(articleContent)
	dbi.Save(group)

	for i:=0; i<30; i++ {
		w := httptest.NewRecorder()
		articleComment, err := tests.CreateArticleComment(article, group, user)
		assert.Equal(t, err, nil)
		dbi.Save(articleComment)

		models.CommentArticleIncentive(articleComment, dbi)

		req, _ := http.NewRequest("GET", "/v1/users/" + user.Address + "/hp", nil)

		router.ServeHTTP(w, req)

		log.Println(w.Body.String())
		assert.Equal(t, w.Code, 200)
	}
}

func TestNonExistUserHP (t *testing.T) {
	tests.InitTestEnv("../../../../config/")
	router := server.NewRouter()
	w := httptest.NewRecorder()

	user, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	req, _ := http.NewRequest("GET", "/v1/users/" + user.Address + "/hp", nil)

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 404)
}