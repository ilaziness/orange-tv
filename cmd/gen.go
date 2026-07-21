package cmd

import (
	"context"
	"fmt"

	"github.com/ilaziness/orange-tv/internal/config"
	"github.com/ilaziness/orange-tv/internal/database"
	"github.com/ilaziness/orange-tv/internal/database/gen"
	"github.com/spf13/cobra"
)

var (
	genTable         string
	genOutput        string
	genPackage       string
	genJSONTags      bool
	genValidatorTags bool
	genWithRelations bool
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Code generation commands",
	Long:  `Code generation commands for generating models, DTOs, and other boilerplate code.`,
}

var genModelCmd = &cobra.Command{
	Use:   "model",
	Short: "Generate model from database",
	Long:  `Generate Go model files from database tables using Bun schema.`,
	RunE:  runGenModel,
}

var genDtoCmd = &cobra.Command{
	Use:   "dto",
	Short: "Generate DTO templates",
	Long:  `Generate DTO (Data Transfer Object) templates for handlers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("dto generation is not yet implemented")
	},
}

func init() {
	genModelCmd.Flags().StringVar(&genTable, "table", "", "Database table name (string; omit to generate all tables)")
	genModelCmd.Flags().StringVar(&genOutput, "output", "./internal/model", "Generated model output directory (string path)")
	genModelCmd.Flags().StringVar(&genPackage, "package", "model", "Generated model package name (string)")
	genModelCmd.Flags().BoolVar(&genJSONTags, "json-tags", true, "Whether to add JSON tags (bool: true or false)")
	genModelCmd.Flags().BoolVar(&genValidatorTags, "validator-tags", false, "Whether to add validator tags (bool: true or false)")
	genModelCmd.Flags().BoolVar(&genWithRelations, "with-relations", false, "Whether to generate relationships from foreign keys (bool: true or false)")

	genCmd.AddCommand(genModelCmd)
	genCmd.AddCommand(genDtoCmd)
	rootCmd.AddCommand(genCmd)
}

func runGenModel(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadWithEnv(cfgFile, env)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logInst, err := newLoggerFromConfig(cfg)
	if err != nil {
		return err
	}

	db, err := database.NewDB(cfg, logInst.Logger)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	ctx := context.Background()

	var tables []string
	if genTable != "" {
		if validateErr := gen.ValidateTableName(genTable); validateErr != nil {
			return fmt.Errorf("invalid --table: %w", validateErr)
		}
		tables = []string{genTable}
	} else {
		tables, err = gen.GetTables(ctx, db, cfg.Database.Driver)
		if err != nil {
			return fmt.Errorf("failed to get tables: %w", err)
		}
	}

	if len(tables) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tables found")
		return nil
	}

	opts := gen.GenerateOptions{
		OutputDir:     genOutput,
		PackageName:   genPackage,
		JSONTags:      genJSONTags,
		ValidatorTags: genValidatorTags,
		WithRelations: genWithRelations,
	}

	if err := gen.GenerateModels(ctx, db, cfg.Database.Driver, tables, opts); err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	for _, tableName := range tables {
		fmt.Fprintf(out, "Generated model for table: %s\n", tableName)
	}
	fmt.Fprintf(out, "\nSuccessfully generated %d model(s)\n", len(tables))

	return nil
}
