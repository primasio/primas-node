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
	"github.com/primasio/primas-node/db"
	"github.com/primasio/primas-node/tests"
	"github.com/magiconair/properties/assert"
	"github.com/ethereum/go-ethereum/crypto"
	"encoding/hex"
	"net/url"
	"strings"
	"strconv"
	"log"
	"time"
	"github.com/primasio/primas-node/models"
)

func TestLikeArticle(t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	dbi := db.GetDb()

	user, privateKey, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)


	group, err := tests.CreateTestGroup(user)
	assert.Equal(t, err, nil)

	article, err := tests.CreateTestArticle(user)
	assert.Equal(t, err, nil)

	groupMember, err := tests.CreateGroupMember(user, group)
	assert.Equal(t, err, nil)

	dbi.Save(&user)
	dbi.Save(&group)
	dbi.Save(&groupMember)
	dbi.Save(&article)

	sigStr := crypto.Keccak256([]byte(article.DNA + group.DNA))

	// Generate signature

	sigBytes, err := crypto.Sign(sigStr, privateKey)

	assert.Equal(t, err, nil)

	signature := hex.EncodeToString(sigBytes)


	data := url.Values{}
	data.Set("GroupDNA", group.DNA)
	data.Set("GroupMemberAddress", user.Address)
	data.Set("Signature", signature)

	req, _ := http.NewRequest("POST", "/v1/articles/" + article.DNA + "/likes", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestCommentArticle(t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	dbi := db.GetDb()

	user, privateKey, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)


	group, err := tests.CreateTestGroup(user)
	assert.Equal(t, err, nil)

	article, err := tests.CreateTestArticle(user)
	assert.Equal(t, err, nil)

	groupMember, err := tests.CreateGroupMember(user, group)
	assert.Equal(t, err, nil)

	dbi.Save(&user)
	dbi.Save(&group)
	dbi.Save(&groupMember)
	dbi.Save(&article)

	content := "This is a very nice article."
	contentHash := hex.EncodeToString(crypto.Keccak256([]byte(content)))

	sigStr := crypto.Keccak256([]byte(article.DNA + group.DNA + contentHash))

	// Generate signature

	sigBytes, err := crypto.Sign(sigStr, privateKey)

	assert.Equal(t, err, nil)

	signature := hex.EncodeToString(sigBytes)


	data := url.Values{}
	data.Set("GroupDNA", group.DNA)
	data.Set("GroupMemberAddress", user.Address)
	data.Set("Content", content)
	data.Set("Signature", signature)

	req, _ := http.NewRequest("POST", "/v1/articles/" + article.DNA + "/comments", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestShareArticle(t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	dbi := db.GetDb()

	user, privateKey, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)


	group1, err := tests.CreateTestGroup(user)
	assert.Equal(t, err, nil)

	group2, err := tests.CreateTestGroup(user)
	assert.Equal(t, err, nil)

	article, err := tests.CreateTestArticle(user)
	assert.Equal(t, err, nil)

	groupMember1, err := tests.CreateGroupMember(user, group1)
	assert.Equal(t, err, nil)

	groupMember2, err := tests.CreateGroupMember(user, group2)
	assert.Equal(t, err, nil)

	dbi.Save(&user)
	dbi.Save(&group1)
	dbi.Save(&group2)
	dbi.Save(&groupMember1)
	dbi.Save(&groupMember2)
	dbi.Save(&article)

	sigBase := article.DNA + group1.DNA + "," + group2.DNA

	sigStr := crypto.Keccak256([]byte(sigBase))

	// Generate signature

	sigBytes, err := crypto.Sign(sigStr, privateKey)

	assert.Equal(t, err, nil)

	signature := hex.EncodeToString(sigBytes)

	data := url.Values{}
	data.Set("GroupMemberAddress", user.Address)
	data.Add("GroupDNAs", group1.DNA)
	data.Add("GroupDNAs", group2.DNA)
	data.Set("Signature", signature)

	req, _ := http.NewRequest("POST", "/v1/articles/" + article.DNA + "/groups", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestGetArticleComments(t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	dbi := db.GetDb()

	user, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)


	group, err := tests.CreateTestGroup(user)
	assert.Equal(t, err, nil)

	article, err := tests.CreateTestArticle(user)
	assert.Equal(t, err, nil)

	groupMember, err := tests.CreateGroupMember(user, group)
	assert.Equal(t, err, nil)

	articleComment, err := tests.CreateArticleComment(article, group, user)

	dbi.Save(user)
	dbi.Save(group)
	dbi.Save(groupMember)
	dbi.Save(article)
	dbi.Save(articleComment)

	start := strconv.FormatInt(time.Now().Unix(), 10)

	req, _ := http.NewRequest("GET", "/v1/articles/" + article.DNA + "/comments?start=" + start, nil)

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestShareArticleContract (t *testing.T) {

	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	dbi := db.GetDb()

	article := &models.Article{ ID: 1 }
	dbi.Where(article).First(article)

	group := &models.Group{ ID: 1 }
	dbi.Where(group).First(group)

	keyStore, account, err := tests.LoadTestAccount(1)
	assert.Equal(t, err, nil)

	sigBase := article.DNA + group.DNA

	sigStr := crypto.Keccak256([]byte(sigBase))

	// Generate signature

	sigBytes, err := keyStore.SignHash(*account, sigStr)

	assert.Equal(t, err, nil)

	signature := hex.EncodeToString(sigBytes)

	data := url.Values{}
	data.Set("GroupMemberAddress", account.Address.Hex())
	data.Add("GroupDNAs", group.DNA)
	data.Set("Signature", signature)

	req, _ := http.NewRequest("POST", "/v1/articles/" + article.DNA + "/groups", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestCommentArticleContract (t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	dbi := db.GetDb()

	keyStore, account, err := tests.LoadTestAccount(1)
	assert.Equal(t, err, nil)

	article := &models.Article{ ID: 1 }
	dbi.Where(article).First(article)

	group := &models.Group{ ID: 1 }
	dbi.Where(group).First(group)

	content := "This is a very nice article."
	contentHash := hex.EncodeToString(crypto.Keccak256([]byte(content)))

	sigStr := crypto.Keccak256([]byte(article.DNA + group.DNA + contentHash))

	// Generate signature

	sigBytes, err := keyStore.SignHash(*account, sigStr)
	assert.Equal(t, err, nil)

	signature := hex.EncodeToString(sigBytes)

	data := url.Values{}
	data.Set("GroupDNA", group.DNA)
	data.Set("GroupMemberAddress", account.Address.Hex())
	data.Set("Content", content)
	data.Set("Signature", signature)

	req, _ := http.NewRequest("POST", "/v1/articles/" + article.DNA + "/comments", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestLikeArticleContract (t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	keyStore, account, err := tests.LoadTestAccount(0)
	assert.Equal(t, err, nil)

	dbi := db.GetDb()

	article := &models.Article{ ID: 1 }
	dbi.Where(article).First(article)

	group := &models.Group{ ID: 1 }
	dbi.Where(group).First(group)

	sigStr := crypto.Keccak256([]byte(article.DNA + group.DNA))

	// Generate signature

	sigBytes, err := keyStore.SignHash(*account, sigStr)
	assert.Equal(t, err, nil)

	signature := hex.EncodeToString(sigBytes)


	data := url.Values{}
	data.Set("GroupDNA", group.DNA)
	data.Set("GroupMemberAddress", account.Address.Hex())
	data.Set("Signature", signature)

	req, _ := http.NewRequest("POST", "/v1/articles/" + article.DNA + "/likes", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}
