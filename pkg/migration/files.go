package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// Tries to parse fileinfo to as a migration file meta
func fileToMigration(path string, file os.FileInfo) (*Single, error) {
	r := regexp.MustCompile("^(\\d)_(.+)[.](up|down)[.]sql$")
	match := r.FindSubmatch([]byte(file.Name()))

	mFilename := string(match[0])
	mIx := string(match[1])
	mLabel := string(match[2])
	mDir := string(match[3])

	m := &Single{}

	// Verify migration ix
	ix, err := strconv.ParseInt(mIx, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("<< Migration file has an incorrect index number: '%s'", mFilename)
	}
	m.ix = int(ix)

	// Verify migration dir
	if mDir == "up" {
		m.dir = DirUp
	} else if mDir == "down" {
		m.dir = DirDown
	} else {
		return nil, fmt.Errorf("<< Migration file does not indicate a correct direction (must be 'up' or 'down'): '%s'", mFilename)
	}

	// Verify label
	if mLabel != "" {
		m.label = mLabel
	} else {
		return nil, fmt.Errorf("<< Migration file must contain a label: '%s'", mFilename)
	}

	// Set path
	m.path = path

	return m, nil
}

// find all files that look like migrations and return metadata
func findInDir(path string) []*Single {
	migrationFiles := make([]*Single, 0)

	// Read migration files
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		// Ignore directories
		if info.IsDir() {
			return nil
		}

		migrationFile, err := fileToMigration(path, info)
		if err != nil {
			fmt.Print(err)
			return nil
		}

		migrationFiles = append(migrationFiles, migrationFile)
		return nil
	})

	return migrationFiles
}

// Converts migration meta files to migrations.
// This also matches the up and down files
func pair(list []*Single) (Map, error) {
	mmap := Map{}
	ixHigh := -1

	// Create maps
	for _, m := range list {
		p := mmap[m.ix]

		// Create if does not exist
		if p == nil {
			p = &Pair{}
		}

		// Assign migration
		if m.dir == DirUp {
			if p.up != nil {
				return nil, fmt.Errorf("Duplicate UP migration found for index %d", m.ix)
			}
			p.up = m
		} else if m.dir == DirDown {
			if p.down != nil {
				return nil, fmt.Errorf("Duplicate DOWN migration found for index %d", m.ix)
			}
			p.down = m
		}

		// Assign new highest
		if ixHigh < m.ix {
			ixHigh = m.ix
		}

		mmap[m.ix] = p
	}

	// Verify order and completeness (up and down exist)
	for ix := 1; ix < ixHigh+1; ix++ {
		m := mmap[ix]

		// Should not be nil
		if m == nil {
			return nil, fmt.Errorf("Migration index is not sequential, missing index %d", ix)
		}

		// Should be a pair
		if m.up == nil {
			return nil, fmt.Errorf("Missing UP migration for index %d", ix)
		}
		if m.down == nil {
			return nil, fmt.Errorf("Missing DOWN migration for index %d", ix)
		}
	}

	return mmap, nil
}
