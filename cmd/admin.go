package cmd

import (
	"context"
	"fmt"

	"github.com/ilaziness/orange-tv/internal/repository"
	adminsvc "github.com/ilaziness/orange-tv/internal/service/admin"
	"github.com/spf13/cobra"
)

var (
	adminUsername string
	adminPassword string
	adminEmail    string
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Administrator management commands",
}

var adminCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a super_admin administrator",
	Long: `Create an enabled super_admin administrator.

Requires the phase-2 migration that seeds the super_admin user group.
Does not print the password. Username is globally unique (including soft-deleted).`,
	RunE: runAdminCreate,
}

func init() {
	adminCreateCmd.Flags().StringVar(&adminUsername, "username", "", "admin username (required, 3-50 chars)")
	adminCreateCmd.Flags().StringVar(&adminPassword, "password", "", "admin password (required, 6-72 chars)")
	adminCreateCmd.Flags().StringVar(&adminEmail, "email", "", "admin email (optional)")
	_ = adminCreateCmd.MarkFlagRequired("username")
	_ = adminCreateCmd.MarkFlagRequired("password")

	adminCmd.AddCommand(adminCreateCmd)
	rootCmd.AddCommand(adminCmd)
}

func runAdminCreate(cmd *cobra.Command, args []string) error {
	db, err := loadDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	adminRepo := repository.NewAdminRepo(db)
	admin, err := adminsvc.CreateAdmin(context.Background(), adminRepo, adminUsername, adminPassword, adminEmail)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created super_admin: id=%d username=%s\n", admin.ID, admin.Username)
	return nil
}
