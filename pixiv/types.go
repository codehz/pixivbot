package pixiv

import (
	"fmt"
	"time"
)

type Urls struct {
	Mini     string `json:"mini"`
	Thumb    string `json:"thumb"`
	Small    string `json:"small"`
	Regular  string `json:"regular"`
	Original string `json:"original"`
}

type Tag struct {
	Tag         string            `json:"tag"`
	UserID      string            `json:"userId,omitempty"`
	UserName    string            `json:"userName,omitempty"`
	Translation map[string]string `json:"translation,omitempty"`
}

func (tag Tag) GetTranslation() string {
	for _, val := range tag.Translation {
		return val
	}
	return ""
}

type Tags struct {
	AuthorId string `json:"authorId"`
	Tags     []Tag  `json:"tags"`
}

type TitleCaptionTranslation struct {
	Title   *string `json:"workTitle"`
	Caption *string `json:"workCaption"`
}
type CardInfo struct {
	Description string `json:"description"`
	Image       string `json:"image"`
	Title       string `json:"title"`
	Type        string `json:"type"`
}
type Meta struct {
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	Canonical         string   `json:"canonical"`
	DescriptionHeader string   `json:"descriptionHeader"`
	Ogp               CardInfo `json:"ogp"`
	Twitter           CardInfo `json:"twitter"`
}
type ExtraData struct {
	Meta Meta `json:"meta"`
}

type IllustData struct {
	IllustID                string                  `json:"illustId"`
	IllustTitle             string                  `json:"illustTitle"`
	IllustComment           string                  `json:"illustComment"`
	ID                      string                  `json:"id"`
	Title                   string                  `json:"title"`
	Description             string                  `json:"description"`
	IllustType              int                     `json:"illustType"`
	CreateDate              time.Time               `json:"createDate"`
	UploadDate              time.Time               `json:"uploadDate"`
	Restrict                int                     `json:"restrict"`
	XRestrict               int                     `json:"xRestrict"`
	Sl                      int                     `json:"sl"`
	Urls                    Urls                    `json:"urls"`
	Tags                    Tags                    `json:"tags"`
	Alt                     string                  `json:"alt"`
	StorableTags            []string                `json:"storableTags"`
	UserID                  string                  `json:"userId"`
	UserName                string                  `json:"userName"`
	UserAccount             string                  `json:"userAccount"`
	LikeData                bool                    `json:"likeData"`
	Width                   int                     `json:"width"`
	Height                  int                     `json:"height"`
	PageCount               int                     `json:"pageCount"`
	BookmarkCount           int                     `json:"bookmarkCount"`
	LikeCount               int                     `json:"likeCount"`
	CommentCount            int                     `json:"commentCount"`
	ResponseCount           int                     `json:"responseCount"`
	ViewCount               int                     `json:"viewCount"`
	IsHowto                 bool                    `json:"isHowto"`
	IsOriginal              bool                    `json:"isOriginal"`
	IsBookmarkable          bool                    `json:"isBookmarkable"`
	ExtraData               ExtraData               `json:"extraData"`
	TitleCaptionTranslation TitleCaptionTranslation `json:"titleCaptionTranslation"`
	IsUnlisted              bool                    `json:"isUnlisted"`
}

type PixivResponse interface {
	GetError() error
}

type IllustResponse struct {
	IsError      bool        `json:"error"`
	ErrorMessage string      `json:"message"`
	Body         *IllustData `json:"body"`
}

func (res IllustResponse) GetError() error {
	if res.IsError {
		return fmt.Errorf("server error: %s", res.ErrorMessage)
	}
	return nil
}

type IllustPage struct {
	Urls   Urls `json:"urls"`
	Width  int  `json:"width"`
	Height int  `json:"height"`
}

type IllustPagesResponse struct {
	IsError      bool         `json:"error"`
	ErrorMessage string       `json:"message"`
	Body         []IllustPage `json:"body"`
}

func (res IllustPagesResponse) GetError() error {
	if res.IsError {
		return fmt.Errorf("server error: %s", res.ErrorMessage)
	}
	return nil
}
