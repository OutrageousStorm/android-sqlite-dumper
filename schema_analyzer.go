package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

// AnalyzeSchema examines SQLite database structure
func AnalyzeSchema(dbPath string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	fmt.Printf("\n📊 SQLite Schema Analysis\n")
	fmt.Printf("Database: %s\n", dbPath)
	fmt.Println("=" + "=" + "=\n")

	// Get all tables
	rows, err := db.Query(`
		SELECT name, sql, rootpage FROM sqlite_master 
		WHERE type='table' AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	tableCount := 0
	totalRows := 0

	for rows.Next() {
		var name, sql string
		var rootpage int
		rows.Scan(&name, &sql, &rootpage)
		tableCount++

		// Count rows in table
		countRow := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", name))
		var count int
		countRow.Scan(&count)
		totalRows += count

		fmt.Printf("📋 Table: %s (%d rows)\n", name, count)

		// Get columns
		colRows, _ := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", name))
		var cid int
		var colName, colType string
		var notnull, dfltValue, pk int
		for colRows.Next() {
			colRows.Scan(&cid, &colName, &colType, &notnull, &dfltValue, &pk)
			pkMark := ""
			if pk > 0 {
				pkMark = " [PK]"
			}
			fmt.Printf("  ├─ %s: %s%s\n", colName, colType, pkMark)
		}
		fmt.Println()
	}

	// Get indexes
	indexRows, _ := db.Query(`
		SELECT name, sql FROM sqlite_master 
		WHERE type='index' AND name NOT LIKE 'sqlite_%'
	`)
	var indexCount int
	for indexRows.Next() {
		indexCount++
	}

	fmt.Printf("\n📈 Summary\n")
	fmt.Printf("  Tables: %d\n", tableCount)
	fmt.Printf("  Rows: %d\n", totalRows)
	fmt.Printf("  Indexes: %d\n", indexCount)

	return nil
}

func main() {
	dbPath := flag.String("db", "", "Path to SQLite database")
	flag.Parse()

	if *dbPath == "" {
		fmt.Println("Usage: schema-analyzer -db /path/to/database.db")
		os.Exit(1)
	}

	if err := AnalyzeSchema(*dbPath); err != nil {
		log.Fatal(err)
	}
}
