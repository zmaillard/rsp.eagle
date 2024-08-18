package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type searchResult struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Annotation string   `json:"annotation"`
	Tags       []string `json:"tags"`
	Star       int      `json:"star"`
}

func hasSign(signId string, allSignIds []string) bool {
	for _, id := range allSignIds {
		if id == signId {
			return true
		}
	}
	return false

}

func rebuildIndex(cfg *config) ([]string, error) {
	var imageIds []string
	limit := 200
	offset := 0

	hasData := true

	for hasData {
		url := fmt.Sprintf("%sapi/item/list?token=%s&limit=%v&offset=%v", cfg.EagleApiUrl, cfg.EagleApiToken, limit, offset)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			return imageIds, err
		}
		defer resp.Body.Close()

		searchResult := struct {
			Status string         `json:"status"`
			Data   []searchResult `json:"data"`
		}{}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return imageIds, err
		}

		err = json.Unmarshal(body, &searchResult)
		if err != nil {
			fmt.Println(err)
			return imageIds, err
		}

		for _, item := range searchResult.Data {
			imageIds = append(imageIds, item.Annotation)
		}

		if len(searchResult.Data) == 0 {
			hasData = false
		} else {
			offset += 1
		}

	}

	return imageIds, nil

}
