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
	"github.com/ethereum/go-ethereum/crypto"
	"time"
	"github.com/primasio/go-base36/base36"
	"errors"
)

type Group struct {
	ID              uint `gorm:"primary_key"`
	CreatedAt       uint
	UserAddress     string `gorm:"size:255" binding:"required"`
	Title           string `gorm:"type:text" binding:"required"`
	Description     string `gorm:"type:text" binding:"required"`
	Signature       string `sql:"-" binding:"required"`
	DNA             string `gorm:"size:255;unique_index"`
	Status          string `gorm:"size:64"`
	TxStatus        int `gorm:"type:int"`
	MemberCount     uint `gorm:"type:int unsigned;default:0"`
	ArticleCount    uint `gorm:"type:int unsigned;default:0"`

	// Relations
	IsMember        *GroupMember  `sql:"-" binding:"-"`
	GroupArticles   []Article  `sql:"-" binding:"-"`
}

type GroupMember struct {
	ID              uint `gorm:"primary_key"`
	CreatedAt       uint
	GroupDNA        string `gorm:"size:255"`
	MemberAddress   string `gorm:"size:255" binding:"required"`
	Signature       string `sql:"-" binding:"required"`
	TxStatus        int `gorm:"type:int"`
}

type GroupArticle struct {
	ID              uint `gorm:"primary_key"`
	CreatedAt       uint
	GroupDNA        string `gorm:"size:255"`
	ArticleDNA      string `gorm:"size:255" binding:"required"`
	MemberAddress   string `gorm:"size:255" binding:"required"`
	Signature       string `sql:"-" binding:"required"`
	TxStatus        int `gorm:"type:int"`
}

func NewGroup(title, description string) *Group {
	group := &Group{}
	group.Title = title
	group.Description = description
	group.CreatedAt = uint(time.Now().Unix())
	group.TxStatus = TxStatusPending

	return group
}

func (group *Group) GenerateDNA () (string, error) {

	if group.Signature == "" {
		return "", errors.New("group signature cannot be empty")
	}

	digest := base36.EncodeBytes(crypto.Keccak256([]byte(group.GetDNABaseString())))

	return digest, nil
}

func (group *Group) GetSignatureBaseString () string {
	return group.Title + group.Description
}

func (group *Group) GetDNABaseString () string {
	return group.Signature
}

func (groupMember *GroupMember) GetSignatureBaseString () string {
	return groupMember.GroupDNA
}

func (groupMember *GroupMember) GetOwnerSignatureBaseString () string {
	return groupMember.GroupDNA + groupMember.MemberAddress
}

func (groupArticle *GroupArticle) GetSignatureBaseString () string {
	return groupArticle.GroupDNA + groupArticle.ArticleDNA
}