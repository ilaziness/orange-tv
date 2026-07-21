package cmd

import (
	"context"
	"fmt"
	"time"

	dbmigrate "github.com/ilaziness/orange-tv/internal/database/migrate"
	"github.com/spf13/cobra"
)

var (
	migrateDir string
	dryRun     bool
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands",
	Long:  `Database migration commands for managing database schema changes.`,
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Run database migrations",
	Long:  `Run all pending database migrations to bring the schema up to date.`,
	RunE:  runMigrateUp,
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback database migrations",
	Long:  `Rollback the last database migration.`,
	RunE:  runMigrateDown,
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  `Display the current status of all database migrations.`,
	RunE:  runMigrateStatus,
}

var migrateCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new migration",
	Long:  `Create a new migration file with up and down SQL.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runMigrateCreate,
}

func init() {
	migrateCmd.PersistentFlags().StringVar(&migrateDir, "dir", "./migrations", "Migration file directory (string path)")
	migrateCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Whether to preview migrations without executing them (bool: true or false)")

	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateCreateCmd)
	rootCmd.AddCommand(migrateCmd)
}

func runMigrateUp(cmd *cobra.Command, args []string) error {
	db, err := loadDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	result, err := dbmigrate.Up(ctx, db, migrateDir, dryRun)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	if result.PendingCount == 0 && result.AppliedCount == 0 {
		fmt.Fprintln(out, "No migrations to run")
		return nil
	}

	if dryRun {
		fmt.Fprintf(out, "Dry run: would apply %d migrations\n", result.PendingCount)
		for _, name := range result.PendingNames {
			fmt.Fprintf(out, "  - %s\n", name)
		}
		return nil
	}

	if result.Group != "" {
		fmt.Fprintf(out, "Migrated to %s\n", result.Group)
	} else if result.AppliedCount > 0 {
		fmt.Fprintf(out, "Applied %d migration(s)\n", result.AppliedCount)
	}

	return nil
}

func runMigrateDown(cmd *cobra.Command, args []string) error {
	db, err := loadDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	result, err := dbmigrate.Down(ctx, db, migrateDir, dryRun)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	if result.RollbackCount == 0 {
		fmt.Fprintln(out, "No migrations to rollback")
		return nil
	}

	if dryRun {
		fmt.Fprintf(out, "Dry run: would rollback the last migration group (%d migrations)\n", result.RollbackCount)
		for _, name := range result.RollbackNames {
			fmt.Fprintf(out, "  - %s\n", name)
		}
		return nil
	}

	if result.Group != "" {
		fmt.Fprintf(out, "Rolled back to %s\n", result.Group)
	}

	return nil
}

func runMigrateStatus(cmd *cobra.Command, args []string) error {
	db, err := loadDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	lines, err := dbmigrate.Status(ctx, db, migrateDir)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	if len(lines) == 0 {
		fmt.Fprintln(out, "No migrations found")
		return nil
	}

	fmt.Fprintln(out, "Migration Status:")
	fmt.Fprintln(out, "================")
	for _, line := range lines {
		status := "PENDING"
		if line.Applied {
			status = fmt.Sprintf("APPLIED (Group %d, %s)", line.GroupID, line.MigratedAt.Format(time.RFC3339))
		}
		fmt.Fprintf(out, "%s: %s\n", line.Name, status)
	}
	fmt.Fprintf(out, "\nTotal migrations: %d\n", len(lines))

	return nil
}

func runMigrateCreate(cmd *cobra.Command, args []string) error {
	upFile, downFile, err := dbmigrate.Create(migrateDir, args[0])
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Created migration files:")
	fmt.Fprintf(out, "  %s\n", upFile)
	fmt.Fprintf(out, "  %s\n", downFile)

	return nil
}
