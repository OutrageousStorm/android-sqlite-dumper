package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbPath := flag.String("db", "", "Path to SQLite database")
	table := flag.String("table", "", "Table name (blank=all tables)")
	format := flag.String("format", "json", "Output format: json, csv, or sql")
	output := flag.String("output", "", "Output file (blank=stdout)")
	flag.Parse()

	if *dbPath == "" {
		log.Fatal("❌ Usage: dumper -db <path> [-table name] [-format json|csv|sql] [-output file]")
	}

	if _, err := os.Stat(*dbPath); err != nil {
		log.Fatalf("❌ Database not found: %s", *dbPath)
	}

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatalf("❌ Failed to open database: %v", err)
	}
	defer db.Close()

	// Get all tables if not specified
	tables := []string{}
	if *table != "" {
		tables = []string{*table}
	} else {
		rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
		if err != nil {
			log.Fatalf("❌ Failed to list tables: %v", err)
		}
		defer rows.Close()
		for rows.Next() {
			var name string
			rows.Scan(&name)
			tables = append(tables, name)
		}
	}

	fmt.Fprintf(os.Stderr, "📋 Found %d table(s)
", len(tables))

	// Prepare output
	var writer *os.File = os.Stdout
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			log.Fatalf("❌ Cannot create output file: %v", err)
		}
		defer f.Close()
		writer = f
	}

	switch *format {
	case "json":
		dumpJSON(db, tables, writer)
	case "csv":
		dumpCSV(db, tables, writer)
	case "sql":
		dumpSQL(db, tables, writer)
	default:
		log.Fatal("❌ Unknown format. Use: json, csv, or sql")
	}

	if *output != "" {
		fmt.Fprintf(os.Stderr, "✅ Dumped to: %s
", *output)
	}
}

func dumpJSON(db *sql.DB, tables []string, writer *os.File) {
	fmt.Fprint(writer, "{
")
	for i, table := range tables {
		fmt.Fprintf(writer, "  "%s": [
", table)
		dumpTableJSON(db, table, writer)
		if i < len(tables)-1 {
			fmt.Fprint(writer, "  ],
")
		} else {
			fmt.Fprint(writer, "  ]
")
		}
	}
	fmt.Fprint(writer, "}
")
}

func dumpTableJSON(db *sql.DB, table string, writer *os.File) {
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", table))
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Cannot query %s: %v
", table, err)
		return
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	first := true
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		for i := range cols {
			vals[i] = new(interface{})
		}
		rows.Scan(vals...)

		if !first {
			fmt.Fprint(writer, ",
")
		}
		fmt.Fprint(writer, "    {")
		for j, col := range cols {
			v := *vals[j].(*interface{})
			fmt.Fprintf(writer, ""%s": %v", col, v)
			if j < len(cols)-1 {
				fmt.Fprint(writer, ", ")
			}
		}
		fmt.Fprint(writer, "}")
		first = false
	}
}

func dumpCSV(db *sql.DB, tables []string, writer *os.File) {
	for _, table := range tables {
		fmt.Fprintf(writer, "# Table: %s
", table)
		rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", table))
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Cannot query %s
", table)
			continue
		}
		cols, _ := rows.Columns()
		fmt.Fprint(writer, strings.Join(cols, ","), "
")

		for rows.Next() {
			vals := make([]interface{}, len(cols))
			for i := range cols {
				vals[i] = new(interface{})
			}
			rows.Scan(vals...)

			for i, v := range vals {
				fmt.Fprint(writer, v)
				if i < len(vals)-1 {
					fmt.Fprint(writer, ",")
				}
			}
			fmt.Fprint(writer, "
")
		}
		rows.Close()
		fmt.Fprint(writer, "
")
	}
}

func dumpSQL(db *sql.DB, tables []string, writer *os.File) {
	for _, table := range tables {
		rows, err := db.Query(fmt.Sprintf("SELECT sql FROM sqlite_master WHERE type='table' AND name='%s'", table))
		if err != nil {
			continue
		}
		var sql string
		rows.Scan(&sql)
		if sql != "" {
			fmt.Fprintf(writer, "%s;

", sql)
		}
		rows.Close()

		dataRows, _ := db.Query(fmt.Sprintf("SELECT * FROM %s", table))
		cols, _ := dataRows.Columns()
		for dataRows.Next() {
			vals := make([]interface{}, len(cols))
			for i := range cols {
				vals[i] = new(interface{})
			}
			dataRows.Scan(vals...)

			fmt.Fprintf(writer, "INSERT INTO %s (%s) VALUES (", table, strings.Join(cols, ","))
			for i, v := range vals {
				fmt.Fprintf(writer, "'%v'", v)
				if i < len(vals)-1 {
					fmt.Fprint(writer, ",")
				}
			}
			fmt.Fprint(writer, ");
")
		}
		dataRows.Close()
		fmt.Fprint(writer, "
")
	}
}
