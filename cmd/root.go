package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	"github.com/copygen/copygen/internal/arguments"
	"github.com/copygen/copygen/internal/config"
	"github.com/copygen/copygen/internal/header"
	"github.com/copygen/copygen/internal/view"
)

const (
	file = ".copygen.yaml"
)

var (
	version string

	dryRun bool
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "copygen",
		Short:   "copygen - Automate copyright headers",
		Version: version, // The version is set during the build by making using of `go build -ldflags`.
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("error: expected exactly one argument, got %d", len(args))
			}

			if _, err := os.Stat(args[0]); os.IsNotExist(err) {
				return fmt.Errorf("error: path %s does not exist", args[0])
			}

			vt := arguments.ViewHuman
			v := view.NewRenderer(vt, &view.View{
				Stream: &view.Stream{
					Writer: os.Stdout,
				},
			})

			v.Render(color.BlueString(fmt.Sprintf("Using \"%s\"\n", file)))
			v.Render("\n")
			v.Render(color.BlueString(fmt.Sprintf("Processing \"%s\"\n", args[0])))

			cfg, err := config.Unmarshal(file)
			if err != nil {
				return err
			}

			hp := header.NewProcessor(cfg, args[0], v, dryRun)
			err = hp.Process()
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "dry run")

	return cmd
}

func setupCobraUsageTemplate(cmd *cobra.Command) {
	cobra.AddTemplateFunc("StyleHeading", color.New(color.FgBlue).SprintFunc())
	usageTemplate := cmd.UsageTemplate()
	usageTemplate = strings.NewReplacer(
		`Usage:`, `{{StyleHeading "Usage:"}}`,
		`Examples:`, `{{StyleHeading "Examples:"}}`,
		`Available Commands:`, `{{StyleHeading "Available Commands:"}}`,
		`Flags:`, `{{StyleHeading "Flags:"}}`,
	).Replace(usageTemplate)
	cmd.SetUsageTemplate(usageTemplate)
}

func Execute() {
	cmd := NewRootCommand()
	setupCobraUsageTemplate(cmd)
	cmd.CompletionOptions.DisableDefaultCmd = true
	if err := cmd.Execute(); err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			for _, e := range merr.Errors {
				fmt.Fprintln(os.Stderr, e)
			}
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
