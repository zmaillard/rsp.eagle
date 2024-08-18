package cmd

import (
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"slices"
)

type tag struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type fullTag struct {
	Id   int     `db:"id"`
	Name string  `db:"name"`
	Slug *string `db:"slug"`
}

func getTags(ctx context.Context, db *pgx.Conn) (map[string]int, error) {
	var tags []tag

	sql := `select id, name from sign.tag`

	err := pgxscan.Select(ctx, db, &tags, sql)
	var tagMap = make(map[string]int)

	if err != nil {
		return tagMap, err
	}

	for _, t := range tags {
		tagMap[t.Name] = t.Id
	}

	return tagMap, err
}

func syncTags(db *pgx.Conn, ctx context.Context, update updates, tags *map[string]int) error {
	// Swap Map
	tagIdMap := make(map[int]string)

	for k, v := range *tags {
		tagIdMap[v] = k
	}

	var highwaySignId int
	getSignSql := "SELECT id FROM sign.highwaysign WHERE imageid = $1"
	err := db.QueryRow(ctx, getSignSql, update.ImageId).Scan(&highwaySignId)
	if err != nil {
		return err
	}
	// Get sign tags where new = 1
	sql := fmt.Sprintf("select ht.tag_id, ht.highwaysign_id from sign.tag_highwaysign ht inner join sign.highwaysign hs on ht.highwaysign_id = hs.id where hs.imageid = %s", update.ImageId)

	var signTags []struct {
		TagId  int `db:"tag_id"`
		SignId int `db:"highwaysign_id"`
	}

	// Delete tags not in Eagle
	err = pgxscan.Select(ctx, db, &signTags, sql)
	if err != nil {
		return err
	}

	for _, dbTag := range signTags {
		_, ok := tagIdMap[dbTag.TagId]
		if !ok {
			_, err := db.Exec(ctx, `delete from sign.tag_highwaysign where tag_id = $1 and highwaysign_id = $2`, dbTag.TagId, dbTag.SignId)
			if err != nil {
				return err
			}
		}
	}

	// Get just tagids
	var dbTagIds []int
	for _, dbTag := range signTags {
		dbTagIds = append(dbTagIds, dbTag.TagId)
	}

	// Add tags not in eagle
	var rows [][]interface{}
	for _, eagleTagName := range update.Tags {
		eagleTagId, ok := (*tags)[eagleTagName]
		if !ok {
			return fmt.Errorf("Tag %s not found", eagleTagName)
		}
		if slices.Index(dbTagIds, eagleTagId) == -1 {
			rows = append(rows, []interface{}{eagleTagId, highwaySignId})
		}
	}

	if len(rows) > 0 {
		fmt.Println(rows)
		copyCount, err := db.CopyFrom(ctx, pgx.Identifier{"sign", "tag_highwaysign"}, []string{"tag_id", "highwaysign_id"}, pgx.CopyFromRows(rows))
		if err != nil {
			return err
		}
		fmt.Println(fmt.Sprintf("Inserted %v rows", copyCount))
	}

	return nil

}
