package pixiv

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetIllust(id int) (*IllustData, error) {
	url := fmt.Sprintf("https://www.pixiv.net/ajax/illust/%d", id)
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
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %e", err)
	}
	var illres IllustResponse
	err = json.Unmarshal(data, &illres)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json")
	}
	if illres.IsError {
		return nil, fmt.Errorf("server error: %s", illres.ErrorMessage)
	}
	return illres.Body, nil
}
