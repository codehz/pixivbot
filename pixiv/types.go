package pixiv

import "fmt"

type IllustImages struct {
	IllustImageWidth  string `json:"illust_image_width"`
	IllustImageHeight string `json:"illust_image_height"`
}
type MangaA struct {
	Page     int    `json:"page"`
	URL      string `json:"url"`
	URLSmall string `json:"url_small"`
	URLBig   string `json:"url_big"`
}
type DisplayTags struct {
	Tag                     string `json:"tag"`
	IsPixpediaArticleExists bool   `json:"is_pixpedia_article_exists"`
	SetByAuthor             bool   `json:"set_by_author"`
	IsLocked                bool   `json:"is_locked"`
	IsDeletable             bool   `json:"is_deletable"`
	Translation             string `json:"translation,omitempty"`
}

type TwitterCard struct {
	Card              string `json:"card"`
	Site              string `json:"site"`
	URL               string `json:"url"`
	Title             string `json:"title"`
	Description       string `json:"description"`
	AppNameIphone     string `json:"app:name:iphone"`
	AppIDIphone       string `json:"app:id:iphone"`
	AppURLIphone      string `json:"app:url:iphone"`
	AppNameIpad       string `json:"app:name:ipad"`
	AppIDIpad         string `json:"app:id:ipad"`
	AppURLIpad        string `json:"app:url:ipad"`
	AppNameGoogleplay string `json:"app:name:googleplay"`
	AppIDGoogleplay   string `json:"app:id:googleplay"`
	AppURLGoogleplay  string `json:"app:url:googleplay"`
	Image             string `json:"image"`
}
type Ogp struct {
	Title       string `json:"title"`
	Type        string `json:"type"`
	Image       string `json:"image"`
	Description string `json:"description"`
}
type Meta struct {
	TwitterCard       TwitterCard `json:"twitter_card"`
	Ogp               Ogp         `json:"ogp"`
	Title             string      `json:"title"`
	Description       string      `json:"description"`
	DescriptionHeader string      `json:"description_header"`
	Canonical         string      `json:"canonical"`
}
type TitleCaptionTranslation struct {
	WorkTitle   interface{} `json:"work_title"`
	WorkCaption interface{} `json:"work_caption"`
}

type UgoiraFrame struct {
	File  string `json:"file"`
	Delay int    `json:"delay"`
}

type UgoiraMeta struct {
	Src      string        `json:"src"`
	MimeType string        `json:"mime_type"`
	Frames   []UgoiraFrame `json:"frames"`
}

type IllustDetails struct {
	URL                     string                  `json:"url"`
	URLThumb                string                  `json:"url_s"`
	URLMini                 string                  `json:"url_ss"`
	URLOriginal             string                  `json:"url_big"`
	URLSmall                string                  `json:"url_placeholder"`
	Tags                    []string                `json:"tags"`
	IllustImages            []IllustImages          `json:"illust_images"`
	MangaA                  []MangaA                `json:"manga_a"`
	DisplayTags             []DisplayTags           `json:"display_tags"`
	TagsEditable            bool                    `json:"tags_editable"`
	BookmarkUserTotal       int                     `json:"bookmark_user_total"`
	UgoiraMeta              UgoiraMeta              `json:"ugoira_meta"`
	ShareText               string                  `json:"share_text"`
	Request                 interface{}             `json:"request"`
	Meta                    Meta                    `json:"meta"`
	IsRated                 bool                    `json:"is_rated"`
	ResponseGet             []interface{}           `json:"response_get"`
	ResponseSend            []interface{}           `json:"response_send"`
	BookmarkID              string                  `json:"bookmark_id"`
	BookmarkRestrict        string                  `json:"bookmark_restrict"`
	TitleCaptionTranslation TitleCaptionTranslation `json:"title_caption_translation"`
	IsMypixiv               bool                    `json:"is_mypixiv"`
	IsPrivate               bool                    `json:"is_private"`
	IsHowto                 bool                    `json:"is_howto"`
	IsOriginal              bool                    `json:"is_original"`
	Alt                     string                  `json:"alt"`
	StorableTags            []string                `json:"storable_tags"`
	UploadTimestamp         int                     `json:"upload_timestamp"`
	ID                      string                  `json:"id"`
	UserID                  string                  `json:"user_id"`
	Title                   string                  `json:"title"`
	Width                   string                  `json:"width"`
	Height                  string                  `json:"height"`
	Restrict                string                  `json:"restrict"`
	XRestrict               string                  `json:"x_restrict"`
	Type                    string                  `json:"type"`
	Sl                      int                     `json:"sl"`
	PageCount               string                  `json:"page_count"`
	Comment                 string                  `json:"comment"`
	RatingCount             string                  `json:"rating_count"`
	RatingView              string                  `json:"rating_view"`
	CommentHTML             string                  `json:"comment_html"`
}

type ProfileImg struct {
	Main string `json:"main"`
}
type ExternalSiteWorksStatus struct {
	Booth    bool `json:"booth"`
	Sketch   bool `json:"sketch"`
	Vroidhub bool `json:"vroidHub"`
}
type FanboxDetails struct {
	UserID               string `json:"user_id"`
	CreatorID            string `json:"creator_id"`
	Description          string `json:"description"`
	HasAdultContent      string `json:"has_adult_content"`
	RegistrationDatetime string `json:"registration_datetime"`
	UpdatedDatetime      string `json:"updated_datetime"`
	CoverImageURL        string `json:"cover_image_url"`
	URL                  string `json:"url"`
}
type AuthorDetails struct {
	UserID                  string                  `json:"user_id"`
	UserStatus              string                  `json:"user_status"`
	UserAccount             string                  `json:"user_account"`
	UserName                string                  `json:"user_name"`
	UserPremium             string                  `json:"user_premium"`
	ProfileImg              ProfileImg              `json:"profile_img"`
	ExternalSiteWorksStatus ExternalSiteWorksStatus `json:"external_site_works_status"`
	FanboxDetails           FanboxDetails           `json:"fanbox_details"`
	AcceptRequest           bool                    `json:"accept_request"`
}

type DetailsApi struct {
	IllustDetails IllustDetails `json:"illust_details"`
	AuthorDetails AuthorDetails `json:"author_details"`
}

type PixivResponse interface {
	GetError() error
}

type DetailsResponse struct {
	IsError      bool        `json:"error"`
	ErrorMessage string      `json:"message"`
	Body         *DetailsApi `json:"body"`
}

func (res DetailsResponse) GetError() error {
	if res.IsError {
		return fmt.Errorf("server error: %s", res.ErrorMessage)
	}
	return nil
}
