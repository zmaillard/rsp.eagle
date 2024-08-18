package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"io/ioutil"
	"net/http"
)

type FolderResult struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func listFolders(cfg *config) (map[string]string, error) {
	var folderMap = make(map[string]string)
	resp, err := http.Get(cfg.EagleApiUrl + "api/folder/list" + "?token=" + cfg.EagleApiToken)
	if err != nil {
		fmt.Println(err)
		return folderMap, err
	}
	defer resp.Body.Close()

	folderResult := struct {
		Status string         `json:"status"`
		Data   []FolderResult `json:"data"`
	}{}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return folderMap, err
	}

	err = json.Unmarshal(body, &folderResult)
	if err != nil {
		fmt.Println(err)
		return folderMap, err
	}

	for _, folder := range folderResult.Data {
		folderMap[folder.Name] = folder.Id
	}

	return folderMap, nil
}

func createFolders(db *pgx.Conn, ctx context.Context, cfg *config) (map[string]string, error) {
	var newFolders = make(map[string]string)
	var states []string
	err := pgxscan.Select(ctx, db, &states, "select distinct aas.name from sign.highwaysign inner join sign.admin_area_state aas on aas.id = highwaysign.admin_area_state_id")
	if err != nil {
		return newFolders, err
	}

	for _, state := range states {
		folderReq := struct {
			FolderName string `json:"folderName"`
		}{FolderName: state}

		folderBody, err := json.Marshal(folderReq)
		if err != nil {
			return newFolders, err
		}
		body := bytes.NewReader(folderBody)
		req, err := http.NewRequest(http.MethodPost, cfg.EagleApiUrl+"api/folder/create"+"?token="+cfg.EagleApiToken, body)
		if err != nil {
			return newFolders, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)

		defer resp.Body.Close()

		folderResult := struct {
			Status string       `json:"status"`
			Data   FolderResult `json:"data"`
		}{}

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return newFolders, err
		}

		err = json.Unmarshal(respBody, &folderResult)
		if err != nil {
			return newFolders, err
		}

		newFolders[state] = folderResult.Data.Id
	}

	return newFolders, nil
}
