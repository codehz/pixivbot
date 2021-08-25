package downloader

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/url"

	tb "gopkg.in/tucnak/telebot.v2"
)

type ImageSource interface {
	GetSmallImage() string
	GetOriginalImage() string
}

type ImageFetcher struct {
	UploadMethod
	Original bool
}

type InlineImageFetcher struct {
	InlineImageSource
	Original bool
}

type InlineImageSource interface {
	TransformURL(source string) (string, error)
}

type UploadMethod interface {
	FromURL(source string) (tb.File, error)
}

func (fetcher ImageFetcher) FetchImage(source ImageSource) (tb.File, error) {
	var base string
	if fetcher.Original {
		base = source.GetOriginalImage()
	} else {
		base = source.GetSmallImage()
	}
	return fetcher.FromURL(base)
}

func (fetcher InlineImageFetcher) GetImageUrl(source ImageSource) (string, error) {
	var base string
	if fetcher.Original {
		base = source.GetOriginalImage()
	} else {
		base = source.GetSmallImage()
	}
	return fetcher.TransformURL(base)
}

type DirectURL struct{}
type ProxiedURL struct {
	ProxyHost string
}
type Download struct{}

func (method DirectURL) TransformURL(source string) (string, error) {
	return source, nil
}

func (method DirectURL) FromURL(source string) (tb.File, error) {
	return tb.FromURL(source), nil
}

func (method ProxiedURL) TransformURL(source string) (string, error) {
	ourl, err := url.Parse(source)
	if err != nil {
		return "", err
	}
	ourl.Host = method.ProxyHost
	return ourl.String(), nil
}

func (method ProxiedURL) FromURL(source string) (tb.File, error) {
	base, err := method.TransformURL(source)
	if err != nil {
		return tb.File{}, err
	}
	return tb.FromURL(base), nil
}

func tryEncodeJpeg(buffer *bytes.Buffer, img image.Image, quality int) ([]byte, error) {
	buffer.Reset()
	err := jpeg.Encode(buffer, img, &jpeg.Options{
		Quality: quality,
	})
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func encodeJpeg(img image.Image) ([]byte, error) {
	var quality int = 100
	var buffer bytes.Buffer
	for {
		temp, err := tryEncodeJpeg(&buffer, img, quality)
		if err != nil {
			return nil, err
		}
		if len(temp) < 10*1048576 {
			return temp, nil
		}
		quality -= 10
	}
}

func pngToJpeg(r io.Reader) ([]byte, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}
	return encodeJpeg(img)
}

func (method Download) FromURL(source string) (tb.File, error) {
	request, err := http.NewRequest("GET", source, nil)
	if err != nil {
		return tb.File{}, err
	}
	request.Header.Add("Referer", "https://www.pixiv.net/")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return tb.File{}, err
	}
	defer response.Body.Close()
	contentType := response.Header.Get("Content-Type")
	if contentType == "image/png" {
		data, err := pngToJpeg(response.Body)
		if err != nil {
			return tb.File{}, err
		}
		return tb.FromReader(bytes.NewReader(data)), nil
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return tb.File{}, err
	}
	return tb.FromReader(bytes.NewReader(data)), nil
}
