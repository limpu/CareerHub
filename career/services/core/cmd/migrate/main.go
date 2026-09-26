package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var migrationsDir string
	var command string

	flag.StringVar(&migrationsDir, "dir", "../../../migrations", "Path to migrations folder")
	flag.StringVar(&command, "cmd", "validate", "Command: up, down, or validate")
	flag.Parse()

	log.Printf("[migrate] Running migration tool: command=%s, dir=%s", command, migrationsDir)

	upFile := filepath.Join(migrationsDir, "000001_initial_schema.up.sql")
	downFile := filepath.Join(migrationsDir, "000001_initial_schema.down.sql")

	switch command {
	case "validate":
		validateMigrationFile("UP", upFile)
		validateMigrationFile("DOWN", downFile)
		log.Println("[migrate] All migration files successfully validated!")

	case "up":
		validateMigrationFile("UP", upFile)
		log.Println("[migrate] UP migration ready for execution against PostgreSQL.")

	case "down":
		validateMigrationFile("DOWN", downFile)
		log.Println("[migrate] DOWN migration rollback ready for execution against PostgreSQL.")

	default:
		log.Fatalf("[migrate] Unknown command: %s. Use up, down, or validate.", command)
	}
}

func validateMigrationFile(kind, path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("[migrate] Failed to read %s migration at %s: %v", kind, path, err)
	}

	lines := strings.Split(string(content), "\n")
	statements := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasSuffix(trimmed, ";") && !strings.HasPrefix(trimmed, "--") {
			statements++
		}
	}

	fmt.Printf("[migrate] %s migration (%s): %d bytes, %d SQL statements identified.\n",
		kind, filepath.Base(path), len(content), statements)
}