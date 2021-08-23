package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/codehz/pixivbot/pixiv"
	"github.com/microcosm-cc/bluemonday"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	INVALID_INPUT         = "无效输入"
	NO_LINK               = "找不到关联群组"
	NO_ADMIN              = "无法读取管理员列表: %e"
	POST_SUCCESS          = "发送成功"
	POST_TO_CHANNEL       = "发送到频道"
	POST_ALBUM_TO_CHANNEL = "发送图集到频道（%d 张）"
	NO_PERMISSION         = "用户没有发送权限"
)

var htmlPolicy bluemonday.Policy
var proxied string

func fixString(input string) string {
	return strings.ReplaceAll(input, "<br />", "\n")
}

func getLinkedChat(bot *tb.Bot, orig *tb.Chat) *tb.Chat {
	chat, err := bot.ChatByID(fmt.Sprintf("%d", orig.ID))
	if err != nil {
		return nil
	}
	chat, err = bot.ChatByID(fmt.Sprintf("%d", chat.LinkedChatID))
	if err != nil {
		return nil
	}
	return chat
}

type titleWithURL struct {
	title string
	url   string
}

type tagData struct {
	titleWithURL
	translation string
}

type extractedInfo struct {
	artwork titleWithURL
	author  titleWithURL
	tags    []tagData
}

func (info titleWithURL) getLink(format string) string {
	if format == "#" {
		return fmt.Sprintf("<a href=\"%s\">#%s</a>", info.url, info.title)
	}
	return fmt.Sprintf("<a href=\"%s\"><%s>%s</%s></a>", info.url, format, info.title, format)
}

func isAscii(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func (info tagData) get() string {
	if info.translation != "" {
		if isAscii(info.translation) {
			return fmt.Sprintf("<a href=\"%s\">#%s</a> <i>%s</i>", info.url, info.title, info.translation)
		}
		return fmt.Sprintf("<a href=\"%s\">#%s</a>", info.url, info.translation)
	}
	return fmt.Sprintf("<a href=\"%s\">#%s</a>", info.url, info.title)
}

func extractPixiv(details *pixiv.DetailsApi) (info extractedInfo) {
	info.artwork.title = details.IllustDetails.Title
	info.artwork.url = "https://www.pixiv.net/artworks/" + details.IllustDetails.ID
	info.author.title = details.AuthorDetails.UserName
	info.author.url = "https://www.pixiv.net/users/" + details.AuthorDetails.UserID
	tags := details.IllustDetails.DisplayTags
	info.tags = make([]tagData, len(tags))
	for i := 0; i < len(tags); i++ {
		tag := tags[i]
		info.tags[i] = tagData{
			titleWithURL: titleWithURL{
				title: tag.Tag,
				url:   fmt.Sprintf("https://www.pixiv.net/tags/%s/artworks", tag.Tag),
			},
			translation: tag.Translation,
		}
	}
	return
}

func getCaption(extracted extractedInfo, details *pixiv.DetailsApi) string {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s %s - %s的插画\n", extracted.tags[0].getLink("#"), extracted.artwork.getLink("b"), extracted.author.getLink("i"))
	fmt.Fprintf(&buffer, "👏 %s ❤️ %d 👁️ %s\n", details.IllustDetails.RatingCount, details.IllustDetails.BookmarkUserTotal, details.IllustDetails.RatingView)
	buffer.WriteString(htmlPolicy.Sanitize(fixString(details.IllustDetails.CommentHTML)))
	buffer.WriteByte('\n')
	for i := 0; i < len(extracted.tags); i++ {
		tag := extracted.tags[i]
		fmt.Fprintf(&buffer, "%s ", tag.get())
	}
	return buffer.String()
}

func proxyURL(original string, details *pixiv.DetailsApi) string {
	if proxied == "" {
		return original
	}
	ourl, err := url.Parse(original)
	if err != nil {
		return original
	}
	ourl.Host = proxied
	return ourl.String()
}

func getPhoto(extracted extractedInfo, details *pixiv.DetailsApi) *tb.Photo {
	return &tb.Photo{File: tb.FromURL(proxyURL(details.IllustDetails.URL, details)), Caption: getCaption(extracted, details)}
}

func getAlbum(id int, extracted extractedInfo, details *pixiv.DetailsApi) (album tb.Album, err error) {
	pages := details.IllustDetails.MangaA
	count := len(pages)
	if count > 10 {
		count = 10
	}
	album = make(tb.Album, count)
	caption := getCaption(extracted, details)
	for i, page := range pages[:count] {
		album[i] = &tb.Photo{File: tb.FromURL(proxyURL(page.URL, details))}
	}
	album[0].(*tb.Photo).Caption = caption
	return
}

func getPhotoResult(extracted extractedInfo, details *pixiv.DetailsApi) (result tb.Result) {
	ourl := proxyURL(details.IllustDetails.URL, details)
	result = &tb.PhotoResult{
		URL:         ourl,
		ParseMode:   tb.ModeHTML,
		ThumbURL:    ourl,
		Description: extracted.author.title,
		Title:       extracted.artwork.title,
		Caption:     getCaption(extracted, details),
	}
	result.SetResultID(details.IllustDetails.ID)
	return
}

func parseIllustUrl(input string) (result int, err error) {
	u, err := url.Parse(input)
	if err != nil {
		return
	}
	if u.Scheme != "https" || (u.Host != "www.pixiv.net" && u.Host != "pixiv.net") {
		return 0, fmt.Errorf("not a pixiv link")
	}
	_, err = fmt.Sscanf(u.Path, "/artworks/%d", &result)
	if err == nil {
		return
	}
	if u.Path == "/member_illust.php" {
		return strconv.Atoi(u.Query().Get("illust_id"))
	}
	err = fmt.Errorf("not a illust link")
	return
}

func parseIllustId(input string) (result int, err error) {
	result, err = strconv.Atoi(input)
	if err == nil {
		return
	}
	result, err = parseIllustUrl(input)
	return
}

func makePixiv(bot *tb.Bot, chat *tb.Chat, id int, reply *tb.Message) (err error) {
	bot.Notify(chat, tb.UploadingPhoto)
	details, err := pixiv.GetDetils(id)
	if err != nil {
		return
	}
	extracted := extractPixiv(details)
	photo := getPhoto(extracted, details)
	channel := getLinkedChat(bot, chat)
	if chat.Type == tb.ChatChannel || chat.Type == tb.ChatChannelPrivate {
		_, err = bot.Send(chat, photo, &tb.SendOptions{
			DisableWebPagePreview: true,
			ParseMode:             "html",
			ReplyTo:               reply,
		})
		return
	}
	menu := &tb.ReplyMarkup{}
	btnArtwork := menu.URL("作品："+extracted.artwork.title, extracted.artwork.url)
	btnAuthor := menu.URL("作者："+extracted.author.title, extracted.author.url)
	btnDownload := menu.URL("下载原图", details.IllustDetails.URLOriginal)
	if channel != nil {
		btnPost := menu.Data(POST_TO_CHANNEL, "post", details.IllustDetails.ID)
		if len(details.IllustDetails.MangaA) > 1 {
			btnPostMulti := menu.Data(fmt.Sprintf(POST_ALBUM_TO_CHANNEL, len(details.IllustDetails.MangaA)), "post-multi", details.IllustDetails.ID)
			menu.Inline(menu.Row(btnPost), menu.Row(btnPostMulti), menu.Row(btnArtwork), menu.Row(btnAuthor), menu.Row(btnDownload))
		} else {
			menu.Inline(menu.Row(btnPost), menu.Row(btnArtwork), menu.Row(btnAuthor), menu.Row(btnDownload))
		}
	} else {
		menu.Inline(menu.Row(btnArtwork), menu.Row(btnAuthor), menu.Row(btnDownload))
	}
	_, err = bot.Send(chat, photo, &tb.SendOptions{
		DisableWebPagePreview: true,
		ParseMode:             "html",
		ReplyTo:               reply,
	}, menu)
	return
}

func makeAlbum(bot *tb.Bot, chat *tb.Chat, id int) (err error) {
	bot.Notify(chat, tb.UploadingPhoto)
	details, err := pixiv.GetDetils(id)
	if err != nil {
		return
	}
	extracted := extractPixiv(details)
	album, err := getAlbum(id, extracted, details)
	if err != nil {
		return
	}
	_, err = bot.SendAlbum(chat, album, &tb.SendOptions{
		DisableWebPagePreview: true,
		ParseMode:             "html",
	})
	return
}

func precheckInlineButton(bot *tb.Bot, c *tb.Callback) (*tb.Chat, error) {
	ochat := c.Message.OriginalChat
	if ochat == nil {
		ochat = c.Message.Chat
	}
	bot.Notify(ochat, tb.Typing)
	linked := getLinkedChat(bot, ochat)
	if linked == nil {
		return nil, fmt.Errorf(NO_LINK)
	}
	members, err := bot.AdminsOf(linked)
	if err != nil {
		return nil, fmt.Errorf(NO_ADMIN, err)
	}
	for _, member := range members {
		if member.User.ID == c.Sender.ID {
			return linked, nil
		}
	}
	return nil, fmt.Errorf(NO_PERMISSION)
}

//go:embed help.txt
var helpMessage string

func main() {
	var token string
	flag.StringVar(&token, "t", "", "Telegram token")
	flag.StringVar(&proxied, "p", "", "i.pximg.net proxy for bypass restrict")
	flag.Parse()
	bot, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	htmlPolicy = *bluemonday.NewPolicy()

	htmlPolicy.AllowStandardURLs()
	htmlPolicy.AllowAttrs("href").OnElements("a")
	htmlPolicy.AllowNoAttrs().OnElements(
		"b", "i", "u", "s",
		"strong", "em", "ins", "strike", "del",
		"code", "pre",
	)

	bot.Handle("/help", func(m *tb.Message) {
		bot.Send(m.Chat, helpMessage, &tb.SendOptions{
			DisableWebPagePreview: true,
			ParseMode:             "html",
			ReplyTo:               m,
		})
	})
	bot.Handle("/start", func(m *tb.Message) {
		bot.Send(m.Chat, helpMessage, &tb.SendOptions{
			DisableWebPagePreview: true,
			ParseMode:             "html",
			ReplyTo:               m,
		})
	})
	bot.Handle("/pixiv", func(m *tb.Message) {
		value, err := parseIllustId(m.Payload)
		if err != nil {
			bot.Send(m.Chat, INVALID_INPUT)
			return
		}
		err = makePixiv(bot, m.Chat, value, nil)
		if err != nil {
			bot.Send(m.Chat, err.Error())
			return
		}
		bot.Delete(m)
	})
	bot.Handle("/album", func(m *tb.Message) {
		value, err := parseIllustId(m.Payload)
		if err != nil {
			bot.Send(m.Chat, INVALID_INPUT)
			return
		}
		err = makeAlbum(bot, m.Chat, value)
		if err != nil {
			bot.Send(m.Chat, err.Error())
			return
		}
		bot.Delete(m)
	})
	bot.Handle("/post", func(m *tb.Message) {
		value, err := parseIllustId(m.Payload)
		if err != nil {
			bot.Send(m.Chat, INVALID_INPUT)
			return
		}
		channel := getLinkedChat(bot, m.Chat)
		if channel == nil {
			bot.Send(m.Chat, NO_LINK)
			return
		}
		err = makePixiv(bot, channel, value, nil)
		if err != nil {
			bot.Send(m.Chat, err.Error())
			return
		}
		bot.Delete(m)
	})
	bot.Handle("/postalbum", func(m *tb.Message) {
		value, err := parseIllustId(m.Payload)
		if err != nil {
			bot.Send(m.Chat, INVALID_INPUT)
			return
		}
		linked := getLinkedChat(bot, m.Chat)
		if linked == nil {
			bot.Send(m.Chat, NO_LINK)
			return
		}
		err = makeAlbum(bot, linked, value)
		if err != nil {
			bot.Send(m.Chat, err.Error())
			return
		}
		bot.Delete(m)
	})
	bot.Handle(&tb.InlineButton{Unique: "post"}, func(c *tb.Callback) {
		linked, err := precheckInlineButton(bot, c)
		if err != nil {
			bot.Respond(c, &tb.CallbackResponse{Text: err.Error(), ShowAlert: true})
			return
		}
		value, err := parseIllustId(c.Data)
		if err != nil {
			bot.Respond(c, &tb.CallbackResponse{Text: INVALID_INPUT + ": " + err.Error(), ShowAlert: true})
			return
		}
		err = makePixiv(bot, linked, value, nil)
		if err != nil {
			bot.Respond(c, &tb.CallbackResponse{Text: err.Error(), ShowAlert: true})
			return
		}
		bot.Respond(c, &tb.CallbackResponse{Text: POST_SUCCESS})
		bot.Delete(c.Message)
	})
	bot.Handle(&tb.InlineButton{Unique: "post-multi"}, func(c *tb.Callback) {
		linked, err := precheckInlineButton(bot, c)
		if err != nil {
			bot.Respond(c, &tb.CallbackResponse{Text: err.Error(), ShowAlert: true})
			return
		}
		value, err := parseIllustId(c.Data)
		if err != nil {
			bot.Respond(c, &tb.CallbackResponse{Text: INVALID_INPUT + ": " + err.Error(), ShowAlert: true})
			return
		}
		err = makeAlbum(bot, linked, value)
		if err != nil {
			bot.Respond(c, &tb.CallbackResponse{Text: err.Error(), ShowAlert: true})
			return
		}
		bot.Respond(c, &tb.CallbackResponse{Text: POST_SUCCESS})
		bot.Delete(c.Message)
	})
	bot.Handle(tb.OnText, func(m *tb.Message) {
		value, err := parseIllustUrl(m.Text)
		if err != nil {
			return
		}
		err = makePixiv(bot, m.Chat, value, m)
		if err != nil {
			bot.Send(m.Chat, err.Error())
			return
		}
	})
	bot.Handle(tb.OnQuery, func(q *tb.Query) {
		value, err := parseIllustId(q.Text)
		if err != nil {
			bot.Answer(q, &tb.QueryResponse{
				Results:      tb.Results{},
				CacheTime:    10,
				SwitchPMText: INVALID_INPUT,
			})
			return
		}
		details, err := pixiv.GetDetils(value)
		if err != nil {
			bot.Answer(q, &tb.QueryResponse{
				Results:      tb.Results{},
				CacheTime:    10,
				SwitchPMText: err.Error(),
			})
			return
		}
		extracted := extractPixiv(details)
		result := getPhotoResult(extracted, details)
		bot.Answer(q, &tb.QueryResponse{
			Results:   tb.Results{result},
			CacheTime: 10,
		})
	})
	bot.Start()
}
