// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sql

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shuLhan/share/lib/reflect"
)

const (
	sqlExtension = ".sql"
)

//
// Client provide a wrapper for generic database instance.
//
type Client struct {
	*sql.DB
	ClientOptions
	TableNames []string // List of tables in database.
}

//
// NewClient create and initialize new database client.
//
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

//
// FetchTableNames return the table names in current database schema sorted
// in ascending order.
//
func (cl *Client) FetchTableNames() (tableNames []string, err error) {
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

	rows, err := cl.DB.Query(q)
	if err != nil {
		return nil, fmt.Errorf("FetchTableNames: " + err.Error())
	}

	if len(cl.TableNames) > 0 {
		cl.TableNames = cl.TableNames[:0]
	}

	for rows.Next() {
		err = rows.Scan(&v)
		if err != nil {
			_ = rows.Close()
			return cl.TableNames, err
		}

		cl.TableNames = append(cl.TableNames, v)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return cl.TableNames, nil
}

//
// Migrate the database using list of SQL files inside a directory.
// Each SQL file in directory will be executed in alphabetical order based on
// the last state.
//
// The state of migration will be saved in table "_migration" including the
// SQL file name that has been executed and the timestamp.
//
func (cl *Client) Migrate(fs http.FileSystem) (err error) {
	if reflect.IsNil(fs) {
		if len(cl.MigrationDir) == 0 {
			return nil
		}
		fs = http.Dir(cl.MigrationDir)
	}

	root, err := fs.Open("/")
	if err != nil {
		return fmt.Errorf("Migrate: %w", err)
	}

	fis, err := root.Readdir(0)
	if err != nil {
		return fmt.Errorf("Migrate: %w", err)
	}

	sort.SliceStable(fis, func(x, y int) bool {
		return fis[x].Name() < fis[y].Name()
	})

	lastFile, err := cl.migrateInit()
	if err != nil {
		return fmt.Errorf("Migrate: %w", err)
	}

	var x int
	if len(lastFile) > 0 {
		for ; x < len(fis); x++ {
			if fis[x].Name() == lastFile {
				break
			}
		}
		if x == len(fis) {
			x = 0
		} else {
			x++
		}
	}
	for ; x < len(fis); x++ {
		name := fis[x].Name()

		sqlRaw, err := loadSQL(fs, fis[x], name)
		if err != nil {
			return fmt.Errorf("Migrate %q: %w", name, err)
		}
		if len(sqlRaw) == 0 {
			continue
		}

		err = cl.migrateApply(name, sqlRaw)
		if err != nil {
			return fmt.Errorf("Migrate %q: %w", name, err)
		}
	}
	return nil
}

//
// migrateInit get the last file in table migration or if its not exist create
// the migration table.
//
func (cl *Client) migrateInit() (lastFile string, err error) {
	lastFile, err = cl.migrateLastFile()
	if err == nil {
		return lastFile, nil
	}

	err = cl.migrateCreateTable()
	if err != nil {
		return "", err
	}

	return "", nil
}

//
// migrateLastFile return the last finished migration or empty string if table
// migration does not exist.
//
func (cl *Client) migrateLastFile() (file string, err error) {
	q := `
		SELECT filename
		FROM _migration
		ORDER BY filename DESC
		LIMIT 1
	`

	err = cl.DB.QueryRow(q).Scan(&file)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	return file, nil
}

func (cl *Client) migrateCreateTable() (err error) {
	q := `
		CREATE TABLE _migration (
			filename    VARCHAR(1024)
		,	applied_at  TIMESTAMP DEFAULT NOW()
		);
	`
	_, err = cl.DB.Exec(q)
	if err != nil {
		return fmt.Errorf("migrateCreateTable: %w", err)
	}
	return nil
}

func (cl *Client) migrateApply(filename string, sqlRaw []byte) (err error) {
	tx, err := cl.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(string(sqlRaw))
	if err == nil {
		err = cl.migrateFinished(tx, filename)
	}
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			log.Printf("migrateApply %s: %s", filename, err2)
		}
		return fmt.Errorf("migrateApply: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("migrateApply: %w", err)
	}

	return nil
}

func (cl *Client) migrateFinished(tx *sql.Tx, file string) (err error) {
	var q string

	switch cl.DriverName {
	case DriverNamePostgres:
		q = `INSERT INTO _migration (filename) VALUES ($1)`
	case DriverNameMysql:
		q = `INSERT INTO _migration (filename) VALUES (?)`
	}

	_, err = tx.Exec(q, file)
	if err != nil {
		return err
	}

	return nil
}

func loadSQL(fs http.FileSystem, fi os.FileInfo, filename string) (
	sqlRaw []byte, err error,
) {
	if strings.ToLower(filepath.Ext(filename)) != sqlExtension {
		return nil, nil
	}

	if !fi.Mode().IsRegular() {
		return nil, nil
	}

	fileSQL := path.Join("/", filename)

	file, err := fs.Open(fileSQL)
	if err != nil {
		return nil, err
	}

	sqlRaw, err = ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return sqlRaw, nil
}

//
// TruncateTable truncate all data on table `tableName`.
//
func (cl *Client) TruncateTable(tableName string) (err error) {
	q := `TRUNCATE TABLE ` + tableName
	_, err = cl.DB.Exec(q)
	if err != nil {
		return fmt.Errorf("TruncateTable %q: %s", tableName, err)
	}
	return nil
}
