package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/pollex/go-migrate/pkg/migration"
	"github.com/urfave/cli/v2"
)

func main() {
	a := &cli.App{
		Name:  "Go Migrate",
		Usage: "Apply database migrations",
		Description: `Using this command the user can apply or undo database migrations.
The second parameter can be used either give a target migration or migrate relative to the current migration.

To migrate to the latest version, omit the second parameter:
	'go-migrate ./migrations'

To migrate to the second migration, supply an integer as second parameter:
	'go-migrate ./migrations 2'

To undo the last 3 migrations, supply a relative integer (prefixed with + or -):
	'go-migrate ./migrations -3'
	`,
		ArgsUsage: "<folder> [(relative) index]",
		Action:    migrate,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "database",
				Aliases: []string{
					"db",
					"d",
				},
				Usage: "Set the connection string of the database to connect to",
			},
		},
	}

	a.Run(os.Args)
}

func migrate(c *cli.Context) error {
	connString := c.String("database")
	if connString == "" {
		fmt.Println("Must provide a database connection string (--database, --db, -d)")
		os.Exit(1)
	}

	path := c.Args().Get(0)
	if path == "" {
		fmt.Println("Command must contain a path to the directory containing the migrations")
		os.Exit(1)
	}
	ixArg := c.Args().Get(1)

	// Create database connection
	pg, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}
	defer pg.Close(context.Background())

	// Create migrator instance
	m, err := migration.NewMigrator(pg, path)
	if err != nil {
		log.Fatalf("Could not create migrator: %s", err)
	}

	// if arg is empty then apply all
	if ixArg == "" {
		err = m.MigrateAll()
	} else {
		ix, err := strconv.ParseInt(ixArg, 10, 32)
		if err != nil {
			log.Fatalf("Incorrect index parameter given.\n")
		}
		// If an argument is given, check whether it is relative
		// if a plus or minus is given then it MUST be relative
		if ixArg[0] == '+' || ixArg[0] == '-' {
			err = m.MigrateRelative(int(ix))
		} else {
			// Otherwise it is an absolute index
			err = m.MigrateTo(int(ix))
		}
	}
	if err != nil {
		log.Fatalf("\nCould not apply migrations: %s", err)
	}

	fmt.Print("Finished applying database migrations!\n")
	return nil
}
