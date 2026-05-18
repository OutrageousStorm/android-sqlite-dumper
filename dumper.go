package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func adb(args ...string) (string, error) {
	cmd := exec.Command("adb", args...)
	out, err := cmd.Output()
	return string(out), err
}

func listAppDatabases(pkg string) ([]string, error) {
	output, err := adb("shell", fmt.Sprintf("find /data/data/%s -name '*.db' 2>/dev/null || find /data/user/0/%s -name '*.db' 2>/dev/null", pkg, pkg))
	if err != nil {
		return nil, err
	}

	var dbs []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		db := strings.TrimSpace(scanner.Text())
		if db != "" {
			dbs = append(dbs, db)
		}
	}
	return dbs, nil
}

func pullDatabase(pkg string, dbPath string, outDir string) error {
	fileName := filepath.Base(dbPath)
	outPath := filepath.Join(outDir, fmt.Sprintf("%s_%s", pkg, fileName))

	fmt.Printf("  Pulling %s -> %s
", dbPath, outPath)
	_, err := adb("pull", dbPath, outPath)
	return err
}

func main() {
	pkg := flag.String("pkg", "", "Package name to extract (required)")
	outDir := flag.String("out", "./android_dbs", "Output directory")
	flag.Parse()

	if *pkg == "" {
		fmt.Println("Usage: ./dumper -pkg com.example.app [-out ./output]")
		os.Exit(1)
	}

	fmt.Printf("\n🗄️  Android SQLite Dumper\n")
	fmt.Println("═════════════════════════════════════")
	fmt.Printf("Package: %s\n", *pkg)
	fmt.Printf("Output: %s\n\n", *outDir)

	// Create output dir
	os.MkdirAll(*outDir, 0755)

	// List databases
	fmt.Println("Finding databases...")
	dbs, err := listAppDatabases(*pkg)
	if err != nil || len(dbs) == 0 {
		fmt.Println("  No databases found or permission denied")
		return
	}

	fmt.Printf("  Found %d database(s)\n\n", len(dbs))

	// Pull each database
	for _, db := range dbs {
		if err := pullDatabase(*pkg, db, *outDir); err != nil {
			fmt.Printf("  ✗ Error: %v\n", err)
		} else {
			fmt.Println("  ✓ Pulled")
		}
	}

	fmt.Printf("\n✅ Extraction complete. Open with: sqlite3 %s\n", filepath.Join(*outDir, "*"))
}
