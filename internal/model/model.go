// Package model: Storing Model definitions
package model

import (
	"strings"

	"github.com/Catizard/bmstable"
	"gorm.io/gorm"
)

type DiffTableHeader struct {
	gorm.Model

	HeaderUrl    string
	DataUrl      string
	Name         string
	OriginalUrl  *string
	Symbol       string
	OrderNumber  int `gorm:"default:0"`
	LevelOrders  string
	TagColor     string
	TagTextColor string
	NoTagBuild   *int `gorm:"default:0"`
}

func (DiffTableHeader) TableName() string {
	return "difftable_header"
}

// Convert external difficult table definition to internal one
// If inheritHeader is non-nil, inherit some extra fields from it (esp color definitions)
func NewDiffTableHeaderFromImport(importHeader *bmstable.DifficultTable, inheritHeader *DiffTableHeader) *DiffTableHeader {
	ret := &DiffTableHeader{
		HeaderUrl:   importHeader.HeaderURL,
		DataUrl:     importHeader.DataURL,
		Name:        importHeader.Name,
		OriginalUrl: &importHeader.OriginalURL,
		Symbol:      importHeader.Symbol,
		LevelOrders: strings.Join(importHeader.LevelOrder, ","),
	}
	if inheritHeader != nil {
		ret.TagColor = inheritHeader.TagColor
		ret.TagTextColor = inheritHeader.TagTextColor
		ret.NoTagBuild = inheritHeader.NoTagBuild
	}
	return ret
}

type DiffTableData struct {
	gorm.Model
	HeaderID uint
	Artist   string
	Comment  string
	Level    string
	Lr2BmsId string `json:"lr2_bmdid"`
	Md5      string `gorm:"index"`
	NameDiff string
	Title    string
	Url      string `json:"url"`
	UrlDiff  string `json:"url_diff"`
	Sha256   string
}

func (DiffTableData) TableName() string {
	return "difftable_data"
}

// Convert bmstable's type definition into internal one
func NewDiffTableDataFromImport(importData *bmstable.DifficultTableData) *DiffTableData {
	return &DiffTableData{
		Artist:   importData.Artist,
		Comment:  importData.Comment,
		Level:    importData.Level,
		Lr2BmsId: importData.Lr2BmsID,
		Md5:      importData.Md5,
		NameDiff: importData.NameDiff,
		Title:    importData.Title,
		Url:      importData.URL,
		UrlDiff:  importData.URLDiff,
		Sha256:   importData.Sha256,
	}
}
