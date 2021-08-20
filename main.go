package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/codehz/pixivbot/pixiv"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	INVALID_INPUT = "无效输入"
	NO_LINK       = "找不到关联群组"
	NO_ADMIN      = "无法读取管理员列表"
	POST_SUCCESS  = "发送成功"
	NO_PERMISSION = "用户没有发送权限"
)

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

func isEnglish(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && r != ' ' && r != '-' {
			return false
		}
	}
	return true
}

func (info tagData) get() string {
	if info.translation != "" {
		if isEnglish(info.translation) {
			return fmt.Sprintf("<a href=\"%s\">#%s</a> <i>%s</i>", info.url, info.title, info.translation)
		}
		return fmt.Sprintf("<a href=\"%s\">#%s</a>", info.url, info.translation)
	}
	return fmt.Sprintf("<a href=\"%s\">#%s</a>", info.url, info.title)
}

func extractPixiv(data *pixiv.IllustData) (info extractedInfo) {
	info.artwork.title = data.Title
	info.artwork.url = "https://www.pixiv.net/artworks/" + data.ID
	info.author.title = data.UserName
	info.author.url = "https://www.pixiv.net/users/" + data.UserID
	tags := data.Tags.Tags
	info.tags = make([]tagData, len(tags))
	for i := 0; i < len(tags); i++ {
		tag := tags[i]
		info.tags[i] = tagData{
			titleWithURL: titleWithURL{
				title: tag.Tag,
				url:   fmt.Sprintf("https://www.pixiv.net/tags/%s/artworks", tag.Tag),
			},
			translation: tag.GetTranslation(),
		}
	}
	return
}

func getPhoto(extracted extractedInfo, illust *pixiv.IllustData) *tb.Photo {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s - %s的插画\n", extracted.artwork.getLink("b"), extracted.author.getLink("i"))
	buffer.WriteString(fixString(illust.IllustComment))
	buffer.WriteByte('\n')
	for i := 0; i < len(extracted.tags); i++ {
		tag := extracted.tags[i]
		fmt.Fprintf(&buffer, "%s ", tag.get())
	}
	return &tb.Photo{File: tb.FromURL(illust.Urls.Regular), Caption: buffer.String()}
}

func getPhotoResult(extracted extractedInfo, illust *pixiv.IllustData) (result tb.Result) {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s - %s的插画\n", extracted.artwork.getLink("b"), extracted.author.getLink("i"))
	buffer.WriteString(fixString(illust.IllustComment))
	buffer.WriteByte('\n')
	for i := 0; i < len(extracted.tags); i++ {
		tag := extracted.tags[i]
		fmt.Fprintf(&buffer, "%s ", tag.get())
	}
	result = &tb.PhotoResult{
		URL:         illust.Urls.Regular,
		ParseMode:   tb.ModeHTML,
		ThumbURL:    illust.Urls.Thumb,
		Description: extracted.author.title,
		Title:       extracted.artwork.title,
		Caption:     buffer.String(),
	}
	result.SetResultID(illust.IllustID)
	return
}

func main() {
	var token string
	flag.StringVar(&token, "t", "", "Telegram token")
	flag.Parse()
	bot, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	bot.Handle("/help", func(m *tb.Message) {
		bot.Send(m.Chat, "WIP")
	})
	bot.Handle("/start", func(m *tb.Message) {
		bot.Send(m.Chat, "WIP")
	})
	bot.Handle("/pixiv", func(m *tb.Message) {
		value, err := strconv.Atoi(m.Payload)
		if err != nil {
			bot.Send(m.Chat, INVALID_INPUT)
			return
		}
		bot.Notify(m.Chat, tb.UploadingPhoto)
		illust, err := pixiv.GetIllust(value)
		if err != nil {
			bot.Send(m.Chat, err.Error())
			return
		}
		extracted := extractPixiv(illust)
		photo := getPhoto(extracted, illust)
		channel := getLinkedChat(bot, m.Chat)
		menu := &tb.ReplyMarkup{}
		btnArtwork := menu.URL("作品："+extracted.artwork.title, extracted.artwork.url)
		btnAuthor := menu.URL("作者："+extracted.author.title, extracted.author.url)
		if channel != nil {
			btnPost := menu.Data("Post to channel", "post", m.Payload)
			menu.Inline(menu.Row(btnPost), menu.Row(btnArtwork), menu.Row(btnAuthor))
		} else {
			menu.Inline(menu.Row(btnArtwork), menu.Row(btnAuthor))
		}
		_, err = bot.Send(m.Chat, photo, &tb.SendOptions{
			DisableWebPagePreview: true,
			ParseMode:             "html",
		}, menu)
		if err != nil {
			log.Printf("Send error: %s", err)
			return
		}
		bot.Delete(m)
	})
	bot.Handle("/post", func(m *tb.Message) {
		value, err := strconv.Atoi(m.Payload)
		if err != nil {
			bot.Send(m.Chat, INVALID_INPUT)
			return
		}
		illust, err := pixiv.GetIllust(value)
		if err != nil {
			bot.Send(m.Chat, err.Error())
			return
		}
		channel := getLinkedChat(bot, m.Chat)
		if channel == nil {
			bot.Send(m.Chat, NO_LINK)
			return
		}
		bot.Notify(channel, tb.UploadingPhoto)
		extracted := extractPixiv(illust)
		photo := getPhoto(extracted, illust)
		_, err = bot.Send(channel, photo, &tb.SendOptions{
			DisableWebPagePreview: true,
			ParseMode:             "html",
		})
		if err != nil {
			log.Printf("Send error: %s", err)
			return
		}
		bot.Delete(m)
	})
	bot.Handle(&tb.InlineButton{Unique: "post"}, func(c *tb.Callback) {
		ochat := c.Message.OriginalChat
		if ochat == nil {
			ochat = c.Message.Chat
		}
		bot.Notify(ochat, tb.Typing)
		linked := getLinkedChat(bot, ochat)
		if linked == nil {
			bot.Respond(c, &tb.CallbackResponse{Text: NO_LINK, ShowAlert: true})
			return
		}
		members, err := bot.AdminsOf(linked)
		if err != nil {
			bot.Respond(c, &tb.CallbackResponse{Text: NO_ADMIN + ": " + err.Error(), ShowAlert: true})
			return
		}
		for _, member := range members {
			if member.User.ID == c.Sender.ID {
				value, err := strconv.Atoi(c.Data)
				if err != nil {
					bot.Respond(c, &tb.CallbackResponse{Text: INVALID_INPUT + ": " + err.Error(), ShowAlert: true})
					return
				}
				illust, err := pixiv.GetIllust(value)
				if err != nil {
					bot.Respond(c, &tb.CallbackResponse{Text: err.Error(), ShowAlert: true})
					return
				}
				extracted := extractPixiv(illust)
				photo := getPhoto(extracted, illust)
				_, err = bot.Send(linked, photo, &tb.SendOptions{
					DisableWebPagePreview: true,
					ParseMode:             "html",
				})
				if err != nil {
					bot.Respond(c, &tb.CallbackResponse{Text: err.Error(), ShowAlert: true})
					return
				}
				bot.Respond(c, &tb.CallbackResponse{Text: POST_SUCCESS})
				bot.Delete(c.Message)
				return
			}
		}
		bot.Respond(c, &tb.CallbackResponse{Text: NO_PERMISSION, ShowAlert: true})
	})
	bot.Handle(tb.OnQuery, func(q *tb.Query) {
		value, err := strconv.Atoi(q.Text)
		if err != nil {
			bot.Answer(q, &tb.QueryResponse{
				Results:      tb.Results{},
				CacheTime:    10,
				SwitchPMText: INVALID_INPUT,
			})
			return
		}
		illust, err := pixiv.GetIllust(value)
		if err != nil {
			bot.Answer(q, &tb.QueryResponse{
				Results:      tb.Results{},
				CacheTime:    10,
				SwitchPMText: err.Error(),
			})
			return
		}
		extracted := extractPixiv(illust)
		result := getPhotoResult(extracted, illust)
		bot.Answer(q, &tb.QueryResponse{
			Results:   tb.Results{result},
			CacheTime: 10,
		})
	})
	bot.Start()
}
