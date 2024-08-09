// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sql

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/memfs"
	"git.sr.ht/~shulhan/pakakeh.go/lib/reflect"
)

const (
	sqlExtension = ".sql"
)

// Client provide a wrapper for generic database instance.
type Client struct {
	*sql.DB
	ClientOptions
	TableNames []string // List of tables in database.
}

// NewClient create and initialize new database client.
func NewClient(opts ClientOptions) (cl *Client, err error) {
	db, err := sql.Open(opts.DriverName, opts.DSN)
	if err != nil {
		return nil, fmt.Errorf("sql.NewClient: %w", err)
	}

	cl = &Client{
		DB:            db,
		ClientOptions: opts,
	}

	return cl, nil
}

// FetchTableNames return the table names in current database schema sorted
// in ascending order.
func (cl *Client) FetchTableNames() (tableNames []string, err error) {
	var logp = `FetchTableNames`
	var q, v string

	switch cl.DriverName {
	case DriverNameMysql, DriverNamePostgres:
		q = `
			SELECT
				table_name
			FROM
				information_schema.tables
			ORDER BY
				table_name
		`
	}

	var rows *sql.Rows

	rows, err = cl.DB.Query(q)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	if len(cl.TableNames) > 0 {
		cl.TableNames = cl.TableNames[:0]
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		err = rows.Scan(&v)
		if err != nil {
			return cl.TableNames, fmt.Errorf(`%s: %w`, logp, err)
		}

		cl.TableNames = append(cl.TableNames, v)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return cl.TableNames, nil
}

// Migrate the database using list of SQL files inside a directory.
// Each SQL file in directory will be executed in alphabetical order based on
// the last state.
//
// The table parameter contains the name of table where the state of migration
// will be saved.
// If its empty default to "_migration".
// The state including the SQL file name that has been executed and the
// timestamp.
func (cl *Client) Migrate(tableMigration string, fs *memfs.MemFS) (err error) {
	logp := "Migrate"

	if reflect.IsNil(fs) {
		if len(cl.MigrationDir) == 0 {
			return nil
		}
		var mfsopts = memfs.Options{
			Root: cl.MigrationDir,
			Includes: []string{
				`.*\.(sql)$`,
			},
		}
		fs, err = memfs.New(&mfsopts)
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	root, err := fs.Open("/")
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	fis, err := root.Readdir(0)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	sort.SliceStable(fis, func(x, y int) bool {
		return fis[x].Name() < fis[y].Name()
	})

	if len(tableMigration) == 0 {
		tableMigration = "_migration"
	}

	lastFile, err := cl.migrateInit(tableMigration)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	var (
		x              int
		lastFileExists bool
	)
	if len(lastFile) > 0 {
		for ; x < len(fis); x++ {
			if fis[x].Name() == lastFile {
				lastFileExists = true
				break
			}
		}
		x++
		// If the last file not found, there will be no SQL script to
		// be executed.  Since we cannot return it as an error, we
		// only log it here.  In the future, we may return it.
		if !lastFileExists {
			log.Printf("%s: the last file %s not found on the list",
				logp, lastFile)
		}
	}
	for ; x < len(fis); x++ {
		name := fis[x].Name()

		sqlRaw, err := loadSQL(fs, fis[x], name)
		if err != nil {
			return fmt.Errorf("%s: %q: %w", logp, name, err)
		}
		if len(sqlRaw) == 0 {
			continue
		}

		err = cl.migrateApply(tableMigration, name, sqlRaw)
		if err != nil {
			return fmt.Errorf("%s: %q: %w", logp, name, err)
		}
	}
	return nil
}

// migrateInit get the last file in table migration or if its not exist create
// the migration table.
func (cl *Client) migrateInit(tableMigration string) (lastFile string, err error) {
	lastFile, err = cl.migrateLastFile(tableMigration)
	if err == nil {
		return lastFile, nil
	}

	err = cl.migrateCreateTable(tableMigration)
	if err != nil {
		return "", fmt.Errorf("migrateInit: %w", err)
	}

	return "", nil
}

// migrateLastFile return the last finished migration or empty string if table
// migration does not exist.
func (cl *Client) migrateLastFile(tableMigration string) (file string, err error) {
	q := `
		SELECT filename
		FROM ` + tableMigration + `
		ORDER BY filename DESC
		LIMIT 1
	`

	err = cl.DB.QueryRow(q).Scan(&file)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("migrateLastFile: %w", err)
	}

	return file, nil
}

func (cl *Client) migrateCreateTable(tableMigration string) (err error) {
	q := `
		CREATE TABLE ` + tableMigration + ` (
			filename    VARCHAR(1024)
		,	applied_at  TIMESTAMP DEFAULT NOW()
		,	PRIMARY KEY(filename)
		);
	`
	_, err = cl.DB.Exec(q)
	if err != nil {
		return fmt.Errorf("migrateCreateTable: %w", err)
	}
	return nil
}

func (cl *Client) migrateApply(tableMigration, filename string, sqlRaw []byte) (err error) {
	var (
		logp = "migrateApply"
		tx   *sql.Tx
	)

	tx, err = cl.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(string(sqlRaw))
	if err == nil {
		err = cl.migrateFinished(tx, tableMigration, filename)
	}
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			log.Printf("%s: %s: %s", logp, filename, err2)
		}
		return fmt.Errorf("%s: %w", logp, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	return nil
}

func (cl *Client) migrateFinished(tx *sql.Tx, tableMigration, file string) (err error) {
	var (
		logp = "migrateFinished"
		q    string
	)

	switch cl.DriverName {
	case DriverNamePostgres:
		q = `INSERT INTO ` + tableMigration + ` (filename) VALUES ($1)`
	case DriverNameMysql:
		q = `INSERT INTO ` + tableMigration + ` (filename) VALUES (?)`
	}

	_, err = tx.Exec(q, file)
	if err != nil {
		return fmt.Errorf("%s: %s: %w", logp, file, err)
	}

	return nil
}

func loadSQL(fs *memfs.MemFS, fi os.FileInfo, filename string) (sqlRaw []byte, err error) {
	logp := "loadSQL"

	if strings.ToLower(filepath.Ext(filename)) != sqlExtension {
		return nil, nil
	}

	if !fi.Mode().IsRegular() {
		return nil, nil
	}

	fileSQL := path.Join("/", filename)

	file, err := fs.Open(fileSQL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	sqlRaw, err = io.ReadAll(file)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("%s: %w", logp, err)
		}
	}

	return sqlRaw, nil
}

// TruncateTable truncate all data on table `tableName` with cascade option.
// On PostgreSQL, any identity columns (for example, serial) will be reset
// back to its initial value.
func (cl *Client) TruncateTable(tableName string) (err error) {
	q := fmt.Sprintf(`TRUNCATE TABLE %s %s CASCADE;`, tableName,
		cl.truncateWithRestartIdentity())

	_, err = cl.DB.Exec(q)
	if err != nil {
		return fmt.Errorf(`TruncateTable %q: %w`, tableName, err)
	}
	return nil
}

func (cl *Client) truncateWithRestartIdentity() string {
	if cl.DriverName == DriverNamePostgres {
		return " RESTART IDENTITY "
	}
	return ""
}
