package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/social-platform/services/core/internal/identity"
)

func main() {
	var email, password, fullName, secret string

	flag.StringVar(&email, "email", "", "Super Admin email address (required)")
	flag.StringVar(&password, "password", "", "Super Admin password (min 8 chars, required)")
	flag.StringVar(&fullName, "name", "Super Administrator", "Super Admin full name")
	flag.StringVar(&secret, "secret", "dev_jwt_secret_change_in_production_min_32_chars", "JWT secret")
	flag.Parse()

	if email == "" || password == "" {
		fmt.Println("Usage: bootstrap -email <email> -password <password> [-name <name>]")
		os.Exit(1)
	}

	store := identity.NewMemoryStore()
	authSvc := identity.NewAuthService(store, secret)

	ctx := context.Background()
	admin, err := authSvc.BootstrapSuperAdmin(ctx, email, password, fullName)
	if err != nil {
		log.Fatalf("[bootstrap] Failed to bootstrap Super Admin: %v", err)
	}

	fmt.Printf("[bootstrap] SUCCESS: Initial Super Admin created!\n")
	fmt.Printf("  ID:        %s\n", admin.ID)
	fmt.Printf("  Email:     %s\n", admin.Email)
	fmt.Printf("  Name:      %s\n", admin.FullName)
	fmt.Printf("  Role:      %s\n", admin.PlatformRole)
	fmt.Printf("[bootstrap] One-time bootstrap path is now permanently disabled.\n")
}