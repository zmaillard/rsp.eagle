package cmd

import (
	"context"
	"github.com/caarlos0/env/v11"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	rootCmd.AddCommand(fixSlugCmd)
}

var fixSlugCmd = &cobra.Command{
	Use:   "fix-slug",
	Short: "Fix missing slugs in tag table",
	Run: func(cmd *cobra.Command, args []string) {
		fixSlug()
	},
}

func fixSlug() {
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
	var tags []fullTag

	sql := `select id, name, slug from sign.tag`

	err = pgxscan.Select(ctx, db, &tags, sql)

	if err != nil {
		log.Fatalln(err)

		return
	}

	for _, t := range tags {
		if t.Slug == nil {
			tagSlug := slug.Make(t.Name)
			_, err := db.Exec(ctx, `update sign.tag set slug = $1 where id = $2`, tagSlug, t.Id)
			if err != nil {
				log.Fatalln(err)

				return
			}

		}
	}

}
