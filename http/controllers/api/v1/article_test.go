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
	"github.com/primasio/primas-node/http/server"
	"net/http/httptest"
	"net/http"
	"github.com/magiconair/properties/assert"
	"log"
	"github.com/primasio/primas-node/db"
	"github.com/primasio/primas-node/tests"
	"github.com/ethereum/go-ethereum/crypto"
	"net/url"
	"strings"
	"strconv"
	"github.com/primasio/primas-node/models"
	"encoding/hex"
	"time"
)

func TestPublishArticle(t *testing.T) {

	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	dbi := db.GetDb()

	models.SetState("CurrentBlockHash", hex.EncodeToString(crypto.Keccak256([]byte(tests.RandString(10)))), dbi)

	// Create test user
	keyStore, account, err := tests.LoadTestAccount(0)

	assert.Equal(t, err, nil)

	title := "This is a final test article"
	content := "<p><img src=\"http://primas.io/test.img\"/>This is the final content of the test article</p>"
	license := "article license"
	extra := "{\"language\":\"zh_cn\",\"category\":\"news,economy\",\"type\":\"article\"}"

	contentHash := hex.EncodeToString(crypto.Keccak256([]byte(content)))
	sigStr := crypto.Keccak256([]byte(title + contentHash + license))

	// Generate signature

	sigBytes, err := keyStore.SignHash(*account, sigStr)
	assert.Equal(t, err, nil)

	signature := hex.EncodeToString(sigBytes)

	data := url.Values{}
	data.Set("Title", title)
	data.Set("Content", content)
	data.Set("License", license)
	data.Set("Extra", extra)
	data.Set("Signature", signature)
	data.Set("UserAddress", account.Address.Hex())

	req, _ := http.NewRequest("POST", "/v1/articles", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestGetArticle (t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	dbi := db.GetDb()

	user, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	article, err := tests.CreateTestArticle(user)
	assert.Equal(t, err, nil)

	articleContent, err := tests.CreateTestArticleContent(article)
	assert.Equal(t, err, nil)

	dbi.Save(&user)
	dbi.Save(&article)
	dbi.Save(articleContent)

	req, _ := http.NewRequest("GET", "/v1/articles/" + article.DNA, nil)

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestGetArticleContent(t *testing.T) {

	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	dbi := db.GetDb()

	user, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	article, err := tests.CreateTestArticle(user)
	assert.Equal(t, err, nil)

	articleContent, err := tests.CreateTestArticleContent(article)
	assert.Equal(t, err, nil)

	dbi.Save(&user)
	dbi.Save(&article)
	dbi.Save(articleContent)

	req2, _ := http.NewRequest("GET", "/v1/articles/" + article.DNA + "/content", nil)

	router.ServeHTTP(w, req2)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestArticleList (t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	owner, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	group, err := tests.CreateTestGroup(owner)
	assert.Equal(t, err, nil)

	user, privateKey, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	groupMember, err := tests.CreateGroupMember(user, group)
	assert.Equal(t, err, nil)

	article, err := tests.CreateTestArticle(user)
	assert.Equal(t, err, nil)

	articleContent, err := tests.CreateTestArticleContent(article)
	assert.Equal(t, err, nil)

	groupArticle, err := tests.CreateGroupArticle(article, group, user)
	assert.Equal(t, err, nil)

	dbi := db.GetDb()

	dbi.Save(owner)
	dbi.Save(group)
	dbi.Save(user)
	dbi.Save(groupMember)
	dbi.Save(article)
	dbi.Save(articleContent)
	dbi.Save(groupArticle)

	timestamp := strconv.FormatInt(time.Now().Unix() + 100, 10)

	sigStr := crypto.Keccak256([]byte(timestamp))

	// Generate signature
	sigBytes, err := crypto.Sign(sigStr, privateKey)
	assert.Equal(t, err, nil)

	signature := hex.EncodeToString(sigBytes)

	req, _ := http.NewRequest("GET", "/v1/articles?address=" + user.Address + "&timestamp=" + timestamp + "&signature=" + signature + "&start=" + timestamp, nil)

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestArticleDiscover (t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	owner, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	article, err := tests.CreateTestArticle(owner)
	assert.Equal(t, err, nil)

	articleContent, err := tests.CreateTestArticleContent(article)
	assert.Equal(t, err, nil)

	group, err := tests.CreateTestGroup(owner)
	assert.Equal(t, err, nil)

	groupMember, err := tests.CreateGroupMember(owner, group)
	assert.Equal(t, err, nil)

	groupArticle, err := tests.CreateGroupArticle(article, group, owner)
	assert.Equal(t, err, nil)

	dbi := db.GetDb()

	dbi.Save(owner)
	dbi.Save(article)
	dbi.Save(articleContent)
	dbi.Save(groupMember)
	dbi.Save(groupArticle)

	req, _ := http.NewRequest("GET", "/v1/discover/articles", nil)

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}
