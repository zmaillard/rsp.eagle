package cmd

import (
	"context"
	"github.com/caarlos0/env/v11"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	rootCmd.AddCommand(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync Signs in Eagle With Database",
	Run: func(cmd *cobra.Command, args []string) {
		updateSigns()
	},
}

func updateSigns() {
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
	defer db.Close(ctx)

	if err != nil {
		log.Fatalln(err)
	}

	tags, err := getTags(ctx, db)

	updatedImages, err := getNew(&cfg)
	// Create new tags
	for _, update := range updatedImages {
		for _, t := range update.Tags {
			if _, ok := tags[t]; !ok {
				tagSlug := slug.Make(t)
				var newId int
				row := db.QueryRow(ctx, `insert into sign.tag (name, slug) values ($1, $2)  RETURNING id`, t, tagSlug)
				err := row.Scan(&newId)
				if err != nil {
					log.Fatalln(err)
				}
				tags[t] = newId
			}
		}
	}

	for _, update := range updatedImages {
		err = syncTags(db, ctx, update, &tags)
		if err != nil {
			log.Fatal(err)
		}

		err = updateQuality(db, update)
		if err != nil {
			log.Fatal(err)

		}

	}

}

func updateQuality(db *pgx.Conn, update updates) error {
	sql := `update sign.highwaysign set quality = $1 where imageid = $2`
	_, err := db.Exec(context.Background(), sql, update.Quality, update.ImageId)

	return err
}

type updates struct {
	ImageId string
	Tags    []string
	Quality int
}
