package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
)

func init() {
	rootCmd.AddCommand(importCmd)
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import signs into Eagle",
	Run: func(cmd *cobra.Command, args []string) {
		importSignsIntoEagle()
	},
}

func importSignsIntoEagle() {
	// Load config
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalln(err)
		return
	}

	dbConn := cfg.GetDatabaseConn()

	// Setup database connection
	ctx := context.Background()
	db, err := pgx.Connect(ctx, dbConn)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close(ctx)

	// List folders
	var folderMap map[string]string
	folderMap, err = listFolders(&cfg)
	if err != nil {
		log.Fatal(err)
		return
	}

	if len(folderMap) == 0 {
		// Create folders
		folderMap, err = createFolders(db, ctx, &cfg)
		if err != nil {
			log.Fatal(err)
			return
		}

	}

	// Add signs
	err = addSigns(db, ctx, &cfg, folderMap)
	if err != nil {
		log.Fatal(err)
		return

	}

}

func addSigns(db *pgx.Conn, ctx context.Context, cfg *config, folders map[string]string) error {
	var signs []Sign
	existing, err := rebuildIndex(cfg)
	existing = append(existing, "383427468935079899") //Garbage sign - need to remove
	sql := `select hs.id, hs.title, hs.imageid::text as imageid, coalesce(taggarray.tags, '[]'::json) as tags, aas.name as state
from sign.highwaysign hs
    left outer join (
    select ths.highwaysign_id, json_agg(t.name) as tags from sign.tag_highwaysign ths inner join sign.tag t on t.id = ths.tag_id
    group by ths.highwaysign_id ) taggarray

        on hs.id = taggarray.highwaysign_id
    inner join sign.admin_area_state aas on aas.id = hs.admin_area_state_id`

	err = pgxscan.Select(ctx, db, &signs, sql)
	if err != nil {
		return err
	}

	fmt.Println(len(existing))

	for _, sign := range signs {
		if hasSign(sign.ImageId, existing) {
			continue
		}
		signReq := sign.BuildRequest(cfg, folders)

		signBody, err := json.Marshal(signReq)
		if err != nil {
			return err
		}
		body := bytes.NewReader(signBody)
		req, err := http.NewRequest(http.MethodPost, cfg.EagleApiUrl+"api/item/addFromPath"+"?token="+cfg.EagleApiToken, body)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)

		defer resp.Body.Close()

		signResult := struct {
			Status string `json:"status"`
		}{}

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(respBody, &signResult)
		if err != nil {
			return err
		}

		if signResult.Status != "success" {
			return fmt.Errorf("Failed to add sign.  Response %s", signResult.Status)
		}

		fmt.Println("Loaded " + sign.BuildPath(cfg))

	}

	return nil
}

func getNew(cfg *config) ([]updates, error) {
	var images []updates
	limit := 200
	offset := 0

	hasData := true

	for hasData {
		url := fmt.Sprintf("%sapi/item/list?token=%s&limit=%v&offset=%v", cfg.EagleApiUrl, cfg.EagleApiToken, limit, offset)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			return images, err
		}
		defer resp.Body.Close()

		searchResult := struct {
			Status string         `json:"status"`
			Data   []searchResult `json:"data"`
		}{}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return images, err
		}

		err = json.Unmarshal(body, &searchResult)
		if err != nil {
			fmt.Println(err)
			return images, err
		}

		for _, item := range searchResult.Data {
			if item.Star > 0 {
				images = append(images, updates{
					ImageId: item.Annotation,
					Tags:    item.Tags,
					Quality: item.Star,
				})
			}
		}

		if len(searchResult.Data) == 0 {
			hasData = false
		} else {
			offset += 1
		}

	}

	return images, nil
}
