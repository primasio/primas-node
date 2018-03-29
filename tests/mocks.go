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

package tests

import (
	"github.com/primasio/primas-node/models"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"crypto/rand"
	"log"
	"time"
	"encoding/hex"
	"github.com/primasio/primas-node/config"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"errors"
	"github.com/ethereum/go-ethereum/accounts"
)

func CreateTestUser() (user *models.User, privateKey *ecdsa.PrivateKey, err error) {

	u := &models.User{}

	privKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)

	if err != nil {
		return nil, nil, err
	}

	u.Name = "Test User " + RandString(5)
	u.Address = crypto.PubkeyToAddress(privKey.PublicKey).Hex()
	u.Extra = "{}"

	log.Println("created test user: " + u.Address)

	return u, privKey, nil
}

func LoadTestAccount(idx int) (*keystore.KeyStore, *accounts.Account, error) {
	c := config.GetConfig()

	nodeKeystore := keystore.NewKeyStore(c.GetString("test_account.keystore_dir"), keystore.LightScryptN, keystore.LightScryptP)

	if len(nodeKeystore.Accounts()) == 0 {
		return nil, nil, errors.New("node account not found")
	}

	nodeAccount := &nodeKeystore.Accounts()[idx]

	err := nodeKeystore.Unlock(*nodeAccount, c.GetString("node_account.passphrase"))

	if err != nil {
		return nil, nil, err
	}

	return nodeKeystore, nodeAccount, nil
}

func CreateTestGroup(user *models.User) (*models.Group, error) {
	group := &models.Group{}
	group.Title = "A test group " + RandString(5)
	group.Description = "This a test group"
	group.Signature = "Fake signature "  + RandString(5)

	group.UserAddress = user.Address

	group.TxStatus = models.TxStatusConfirmed
	group.CreatedAt = uint(time.Now().Unix())

	var err error
	group.DNA, err = group.GenerateDNA()

	if err != nil {
		return nil, err
	}

	return group, nil
}

func CreateGroupMember(user *models.User, group *models.Group) (*models.GroupMember, error) {

	groupMember := &models.GroupMember{}
	groupMember.GroupDNA = group.DNA
	groupMember.MemberAddress = user.Address
	groupMember.TxStatus = models.TxStatusConfirmed
	groupMember.CreatedAt = uint(time.Now().Unix())

	return groupMember, nil
}

func CreateTestArticle(user *models.User) (*models.Article, error) {

	article := &models.Article{}

	article.Title = "A test article " + RandString(5)
	article.Content = "This is the content of the test article"
	article.CreatedAt = uint(time.Now().Unix())
	article.UserAddress = user.Address
	article.Extra = "{}"
	article.BlockHash = hex.EncodeToString(crypto.Keccak256([]byte("A1b2c3df")))
	article.License = "{}"
	article.Signature = "TESTSIGNATURE" + RandString(5)

	if abs, err := article.GenerateAbstract(); err == nil {
		article.Abstract = abs
	}else {
		return nil, err
	}

	if cHash, err := article.GenerateContentHash(); err == nil {
		article.ContentHash = cHash
	} else {
		return nil, err
	}

	if dna, err := article.GenerateDNA(); err == nil {
		article.DNA = dna
	} else {
		return nil, err
	}

	return article, nil
}

func CreateTestArticleContent(article *models.Article) (*models.ArticleContent, error) {

	articleContent := &models.ArticleContent{}

	articleContent.DNA = article.DNA
	articleContent.Content = article.Content
	articleContent.CreatedAt = uint(time.Now().Unix())

	return articleContent, nil
}

func CreateGroupArticle (article *models.Article, group *models.Group, user *models.User) (*models.GroupArticle, error) {
	groupArticle := &models.GroupArticle{}
	groupArticle.GroupDNA = group.DNA
	groupArticle.ArticleDNA = article.DNA
	groupArticle.MemberAddress = user.Address
	groupArticle.CreatedAt = uint(time.Now().Unix())

	return groupArticle, nil
}

func CreateArticleComment(article *models.Article, group *models.Group, user *models.User) (*models.ArticleComment, error) {
	articleComment := &models.ArticleComment{}
	articleComment.ArticleDNA = article.DNA
	articleComment.GroupDNA = group.DNA
	articleComment.GroupMemberAddress = user.Address
	articleComment.Content = "This is a very nice article"

	if hash, err := articleComment.GenerateContentHash(); err != nil {
		return nil, err
	} else {
		articleComment.ContentHash = hash
	}

	articleComment.CreatedAt = uint(time.Now().Unix())
	articleComment.TxStatus = models.TxStatusConfirmed

	return articleComment, nil
}

func CreateArticleLike(article *models.Article, group *models.Group, user *models.User) (*models.ArticleLike, error) {
	return &models.ArticleLike{
		ArticleDNA: article.DNA,
		GroupMemberAddress: user.Address,
		GroupDNA: group.DNA,
		TxStatus: models.TxStatusConfirmed,
		CreatedAt: uint(time.Now().Unix()) },
		nil
}