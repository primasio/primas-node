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

import (
	"github.com/jinzhu/gorm"
	"errors"
	"time"
	"github.com/grokify/html-strip-tags-go"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/primasio/go-base36/base36"
	"encoding/hex"
	"github.com/shopspring/decimal"
)

type ArticleContent struct {
	ID             uint `gorm:"primary_key"`
	CreatedAt      uint
	DNA            string `gorm:"size:255;unique_index"`
	Content        string `gorm:"type:longtext"`
}

type Article struct {
	ID              uint `gorm:"primary_key"`
	CreatedAt       uint
	UserAddress     string `gorm:"size:255" binding:"required"`
	Title     		string `gorm:"type:text" binding:"required"`
	Abstract  		string `gorm:"type:text"`
	Content   		string `sql:"-" binding:"required"`
	ContentHash     string `gorm:"size:255"`
	BlockHash       string `gorm:"size:255"`
	DNA       		string `gorm:"size:255;unique_index"`
	License         string `gorm:"type:text" binding:"required"`
	Extra           string `gorm:"type:text" binding:"required"`
	Signature       string `sql:"-" binding:"required"`
	Status          string `gorm:"size:64"`
	TxStatus        int    `gorm:"type:int"`
	LikeCount       uint   `gorm:"default:0"`
	CommentCount    uint   `gorm:"default:0"`
	ShareCount      uint   `gorm:"default:0"`
	TotalIncentives decimal.Decimal `gorm:"type:decimal(65);default:0"`

	// Relations
	Author            User   `gorm:"ForeignKey:UserAddress;AssociationForeignKey:Address" binding:"-"`

	// Temp fields
	GroupDNA        string `sql:"-" binding:"-"`
}

func NewArticleContent (DNA, Content string) *ArticleContent {

	articleContent := &ArticleContent{}
	articleContent.DNA = DNA
	articleContent.Content = Content
	articleContent.CreatedAt = uint(time.Now().Unix())

	return articleContent
}

func (article *Article) GenerateAbstract () (string, error) {

	if article.Content == "" {
		return "", errors.New("article content cannot be empty")
	}

	text := strip.StripTags(article.Content)

	r := []rune(text)

	var cut int

	if len(r) > 160 {
		cut = 160
	} else {
		cut = len(r)
	}

	return string(r[:cut]), nil
}

func (article *Article) GenerateContentHash () (string, error) {

	if article.Content == "" {
		return "", errors.New("article content cannot be empty")
	}

	digest := crypto.Keccak256([]byte(article.Content))

	return hex.EncodeToString(digest), nil
}

func (article *Article) GenerateDNA () (string, error) {

	if article.Signature == "" {
		return "", errors.New("article signature cannot be empty")
	}

	if article.BlockHash == "" {
		return "", errors.New("article block hash cannot be empty")
	}

	digest := base36.EncodeBytes(crypto.Keccak256([]byte(article.GetDNABaseString())))

	return digest, nil
}

func (article *Article) DetectLanguage () (string, error) {
	return "zh-CN", nil
}

func (article *Article) ToMetadata () (title, contentHash, license, blockHash, extras, signature, DNA []byte, err error) {

	signatureBytes, err := hex.DecodeString(article.Signature)

	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	return []byte(article.Title), []byte(article.ContentHash), []byte(article.License), []byte(article.BlockHash), []byte(article.Extra), signatureBytes, []byte(article.DNA), nil
}

func (article *Article) FromMetadata (title, contentHash, license, blockHash, extras, signature, DNA []byte) error {

	article.Title = string(title)
	article.ContentHash = string(contentHash)
	article.License = string(license)
	article.BlockHash = string(blockHash)
	article.Extra = string(extras)
	article.Signature = hex.EncodeToString(signature)
	article.DNA = string(DNA)

	return nil
}

func (article *Article) GetUserAddress () string {
	return article.UserAddress
}

func (article *Article) GetDNA () string {
	return article.DNA
}

func (article *Article) GetContent (db *gorm.DB) (string, error) {

	articleContent := &ArticleContent{}

	db.Model(&article).Related(&articleContent)

	return articleContent.Content, nil
}

func (article *Article) GetSignatureBaseString () string {
	return article.Title + article.ContentHash + article.License
}

func (article *Article) GetDNABaseString () string {
	return article.Signature + article.BlockHash
}