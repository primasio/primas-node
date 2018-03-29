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
	"encoding/hex"
	"errors"
	"github.com/ethereum/go-ethereum/crypto"
	"strings"
)

type ArticleLike struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt uint
	ArticleDNA      string `gorm:"size:255"`
	GroupDNA        string `gorm:"size:255" binding:"required"`
	GroupMemberAddress   string `gorm:"size:255" binding:"required"`
	Signature       string `sql:"-" binding:"required"`
	TxStatus        int `gorm:"type:int"`
}

type ArticleComment struct {
	ID                   uint `gorm:"primary_key"`
	CreatedAt            uint
	GroupDNA             string `gorm:"size:255"`
	GroupMemberAddress   string `gorm:"size:255" binding:"required"`
	ArticleDNA           string `gorm:"size:255"`
	Content   		     string `gorm:"type:longtext" binding:"required"`
	ContentHash          string `gorm:"size:255"`
	Signature            string `sql:"-" binding:"required"`
	TxStatus             int `gorm:"type:int"`

	// Relations
	Author               User  `gorm:"ForeignKey:GroupMemberAddress;AssociationForeignKey:Address" binding:"-"`
}

type ArticleShareBatch struct {
	GroupMemberAddress     string `binding:"required"`
	ArticleDNA             string `gorm:"size:255"`
	GroupDNAs              []string `binding:"required"`
	Signature              string `binding:"required"`
}

func (like *ArticleLike) GetSignatureBaseString() string {
	return like.ArticleDNA + like.GroupDNA
}

func (comment *ArticleComment) GetSignatureBaseString() string {
	return comment.ArticleDNA + comment.GroupDNA + comment.ContentHash
}

func (comment *ArticleComment) GenerateContentHash () (string, error) {

	if comment.Content == "" {
		return "", errors.New("comment content cannot be empty")
	}

	digest := crypto.Keccak256([]byte(comment.Content))

	return hex.EncodeToString(digest), nil
}

func (shareBatch *ArticleShareBatch) GetSignatureBaseString() string {
	return shareBatch.ArticleDNA + shareBatch.GetConcatenatedGroupDNAString()
}

func (shareBatch *ArticleShareBatch) GetConcatenatedGroupDNAString() string {
	return strings.Join(shareBatch.GroupDNAs, ",")
}

func (shareBatch *ArticleShareBatch) FromConcatenatedGroupDNA(groupsDNA string) {
	shareBatch.GroupDNAs = strings.Split(groupsDNA, ",")
}
