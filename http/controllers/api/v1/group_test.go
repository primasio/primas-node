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
	"encoding/hex"
	"github.com/magiconair/properties/assert"
	"net/url"
	"strings"
	"github.com/primasio/primas-node/tests"
	"github.com/ethereum/go-ethereum/crypto"
	"net/http/httptest"
	"net/http"
	"strconv"
	"log"
	"github.com/primasio/primas-node/db"
	"github.com/primasio/primas-node/models"
	"time"
)

func TestGroupCreation (t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	// Create test user
	keyStore, account, err := tests.LoadTestAccount(0)
	assert.Equal(t, err, nil)

	title := "This is a test group"
	description := "This group is test only"

	sigStr := crypto.Keccak256([]byte(title + description))

	// Generate signature
	sigBytes, err := keyStore.SignHash(*account, sigStr)

	assert.Equal(t, err, nil)

	signature := hex.EncodeToString(sigBytes)

	data := url.Values{}
	data.Set("UserAddress", account.Address.Hex())
	data.Set("Title", title)
	data.Set("Description", description)
	data.Set("Signature", signature)

	req, _ := http.NewRequest("POST", "/v1/groups", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestAddGroupMember (t *testing.T) {

	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	dbi := db.GetDb()

	group := &models.Group{ ID: 1 }

	dbi.Where(group).First(group)


	keyStore, account, err := tests.LoadTestAccount(1)
	assert.Equal(t, err, nil)

	user := &models.User{ Address: account.Address.Hex() }

	sigStr := crypto.Keccak256([]byte(group.DNA))

	// Generate signature
	sigBytes, err := keyStore.SignHash(*account, sigStr)

	assert.Equal(t, err, nil)

	signature := hex.EncodeToString(sigBytes)

	data := url.Values{}
	data.Set("MemberAddress", user.Address)
	data.Set("Signature", signature)

	req, _ := http.NewRequest("POST", "/v1/groups/" + group.DNA + "/members", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestRemoveGroupMember (t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	user, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)


	group, err := tests.CreateTestGroup(user)
	assert.Equal(t, err, nil)

	dbi := db.GetDb()

	dbi.Save(&user)
	dbi.Save(&group)

	user2, privateKey, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	groupMember := &models.GroupMember{}
	groupMember.GroupDNA = group.DNA
	groupMember.MemberAddress = user2.Address
	groupMember.CreatedAt = uint(time.Now().Unix())
	groupMember.TxStatus = models.TxStatusConfirmed

	dbi.Save(&groupMember)


	sigStr := crypto.Keccak256([]byte(group.DNA))

	// Generate signature
	sigBytes, err := crypto.Sign(sigStr, privateKey)

	assert.Equal(t, err, nil)

	signature := hex.EncodeToString(sigBytes)

	data := url.Values{}
	data.Set("MemberAddress", user2.Address)
	data.Set("Signature", signature)

	req, _ := http.NewRequest("POST", "/v1/groups/" + group.DNA + "/members/delete", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestRemoveGroupMemberByOwner (t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	user, ownerPrivateKey, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)


	group, err := tests.CreateTestGroup(user)
	assert.Equal(t, err, nil)

	dbi := db.GetDb()

	dbi.Save(&user)
	dbi.Save(&group)

	user2, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	groupMember := &models.GroupMember{}
	groupMember.GroupDNA = group.DNA
	groupMember.MemberAddress = user2.Address
	groupMember.CreatedAt = uint(time.Now().Unix())
	groupMember.TxStatus = models.TxStatusConfirmed

	dbi.Save(&groupMember)


	sigStr := crypto.Keccak256([]byte(group.DNA + groupMember.MemberAddress))

	// Generate signature
	sigBytes, err := crypto.Sign(sigStr, ownerPrivateKey)

	assert.Equal(t, err, nil)

	signature := hex.EncodeToString(sigBytes)

	data := url.Values{}
	data.Set("MemberAddress", user2.Address)
	data.Set("Signature", signature)

	req, _ := http.NewRequest("POST", "/v1/groups/" + group.DNA + "/members/delete/owner", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestListGroup (t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	owner, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	group1, err := tests.CreateTestGroup(owner)
	assert.Equal(t, err, nil)

	group2, err := tests.CreateTestGroup(owner)
	assert.Equal(t, err, nil)

	user, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	group1Member, err := tests.CreateGroupMember(user, group1)
	assert.Equal(t, err, nil)

	group2Member, err := tests.CreateGroupMember(user, group2)
	assert.Equal(t, err, nil)

	dbi := db.GetDb()

	dbi.Save(owner)
	dbi.Save(group1)
	dbi.Save(group2)
	dbi.Save(user)
	dbi.Save(group1Member)
	dbi.Save(group2Member)

	req, _ := http.NewRequest("GET", "/v1/groups?address=" + user.Address, nil)

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestGetGroupArticles(t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	owner, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	group, err := tests.CreateTestGroup(owner)
	assert.Equal(t, err, nil)

	user, _, err := tests.CreateTestUser()
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

	startNum := strconv.FormatInt(time.Now().Unix() + 100, 10)

	req, _ := http.NewRequest("GET", "/v1/groups/" + group.DNA + "/articles?start=" + startNum, nil)

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestGroupDiscover (t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	owner, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	group, err := tests.CreateTestGroup(owner)
	assert.Equal(t, err, nil)

	dbi := db.GetDb()

	dbi.Save(owner)
	dbi.Save(group)

	req, _ := http.NewRequest("GET", "/v1/discover/groups", nil)

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestGetGroupMembers(t *testing.T) {
	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	owner, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	group, err := tests.CreateTestGroup(owner)
	assert.Equal(t, err, nil)

	user1, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	groupMember1, err := tests.CreateGroupMember(user1, group)
	assert.Equal(t, err, nil)

	user2, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	groupMember2, err := tests.CreateGroupMember(user2, group)
	assert.Equal(t, err, nil)

	dbi := db.GetDb()

	dbi.Save(owner)
	dbi.Save(group)
	dbi.Save(user1)
	dbi.Save(user2)
	dbi.Save(groupMember1)
	dbi.Save(groupMember2)

	startNum := strconv.FormatInt(time.Now().Unix() + 100, 10)

	req, _ := http.NewRequest("GET", "/v1/groups/" + group.DNA + "/members?start=" + startNum, nil)

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)
}

func TestGetGroup(t *testing.T) {

	tests.InitTestEnv("../../../../config/")

	router := server.NewRouter()

	w := httptest.NewRecorder()

	owner, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	group1, err := tests.CreateTestGroup(owner)
	assert.Equal(t, err, nil)

	group2, err := tests.CreateTestGroup(owner)
	assert.Equal(t, err, nil)

	user, _, err := tests.CreateTestUser()
	assert.Equal(t, err, nil)

	groupMember1, err := tests.CreateGroupMember(user, group1)
	assert.Equal(t, err, nil)


	article, err := tests.CreateTestArticle(user)
	assert.Equal(t, err, nil)

	articleContent, err := tests.CreateTestArticleContent(article)
	assert.Equal(t, err, nil)


	groupArticle1, err := tests.CreateGroupArticle(article, group1, user)
	assert.Equal(t, err, nil)

	groupArticle2, err := tests.CreateGroupArticle(article, group2, user)
	assert.Equal(t, err, nil)


	dbi := db.GetDb()

	dbi.Save(owner)
	dbi.Save(user)
	dbi.Save(group1)
	dbi.Save(group2)
	dbi.Save(groupMember1)

	dbi.Save(article)
	dbi.Save(articleContent)
	dbi.Save(groupArticle1)
	dbi.Save(groupArticle2)

	req, _ := http.NewRequest("GET", "/v1/groups/" + group1.DNA + "?address=" + user.Address, nil)

	router.ServeHTTP(w, req)

	log.Println(w.Body.String())
	assert.Equal(t, w.Code, 200)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/v1/groups/" + group2.DNA + "?address=" + user.Address, nil)
	router.ServeHTTP(w2, req2)

	log.Println(w2.Body.String())
	assert.Equal(t, w.Code, 200)
}
