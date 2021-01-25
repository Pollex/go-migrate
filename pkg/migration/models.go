package migration

import (
	"io/ioutil"

	"github.com/jackc/pgx/v4"
)

type migrationDirection int

const (
	// DirUp indicates up migration
	DirUp migrationDirection = 1
	// DirDown indicates down migration
	DirDown = 0
)

// Map holds sets of up&down migrations
type Map map[int]*Pair

// Single is a single migration file
type Single struct {
	path  string
	ix    int
	dir   migrationDirection
	label string
}

// Pair holds the respective up and down migration
type Pair struct {
	up   *Single
	down *Single
}

// Migrator is responsible for executing the migrations
type Migrator struct {
	pg         *pgx.Conn
	path       string
	ix         int
	migrations Map
}

// ----- Model functions -----

func (m *Single) readSQL() (string, error) {
	data, err := ioutil.ReadFile(m.path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
