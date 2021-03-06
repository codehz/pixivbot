package pixiv

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func buildRequest(url string) (data []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %e", err)
	}
	req.Header.Set("accept-language", "zh-CN,zh")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request url: %e", err)
	}
	defer response.Body.Close()
	data, err = io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %e", err)
	}
	return
}

func decodeResponse(res PixivResponse, data []byte) error {
	err := json.Unmarshal(data, &res)
	if err != nil {
		println(string(data))
		return fmt.Errorf("failed to parse json")
	}
	return res.GetError()
}

func GetDetils(id int) (*DetailsApi, error) {
	url := fmt.Sprintf("https://www.pixiv.net/touch/ajax/illust/details?illust_id=%d", id)
	data, err := buildRequest(url)
	if err != nil {
		return nil, err
	}
	var details DetailsResponse
	err = decodeResponse(&details, data)
	return details.Body, err
}
