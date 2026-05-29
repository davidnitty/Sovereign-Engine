package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/the-engine/internal/app"
	"github.com/yourusername/the-engine/internal/composition"
	"github.com/yourusername/the-engine/internal/domain"
)

type Options struct {
	CompositionDir string
	Out            io.Writer
	Err            io.Writer
}

func NewRootCommand(opts Options) *cobra.Command {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	if opts.Err == nil {
		opts.Err = os.Stderr
	}
	if opts.CompositionDir == "" {
		opts.CompositionDir = envOrDefault("ENGINE_COMPOSITION_DIR", "compositions")
	}
	cmd := &cobra.Command{
		Use:           "engine",
		Short:         "Sovereign Engine multi-cloud control plane",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetOut(opts.Out)
	cmd.SetErr(opts.Err)
	cmd.AddCommand(
		newDeployCommand(opts),
		newListCommand(opts),
		newStatusCommand(opts),
		newDestroyCommand(opts),
		newCleanupCommand(opts),
		newEncryptCommand(opts),
		newDecryptCommand(opts),
	)
	return cmd
}

func newDeployCommand(opts Options) *cobra.Command {
	var providerName, name, region, size string
	var labels []string
	cmd := &cobra.Command{
		Use:   "deploy <composition>",
		Short: "Deploy a composition",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := app.New(opts.CompositionDir)
			if err != nil {
				return err
			}
			defer a.Close()
			parsedLabels, err := parseLabels(labels)
			if err != nil {
				return err
			}
			resource, err := a.Engine.Deploy(composition.DeployOptions{
				Composition: args[0],
				Provider:    providerName,
				Name:        name,
				Region:      region,
				Size:        size,
				Labels:      parsedLabels,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "deployed %s (%s) on %s\n", resource.Name, resource.ID, resource.Provider)
			return nil
		},
	}
	cmd.Flags().StringVar(&providerName, "provider", "", "cloud provider: aws, azure, gcp, do, hetzner, ovh")
	cmd.Flags().StringVar(&name, "name", "", "resource instance name")
	cmd.Flags().StringVar(&region, "region", "", "provider region")
	cmd.Flags().StringVar(&size, "size", "", "provider size class")
	cmd.Flags().StringArrayVar(&labels, "label", nil, "resource label in key=value form")
	_ = cmd.MarkFlagRequired("provider")
	return cmd
}

func newListCommand(opts Options) *cobra.Command {
	var providerName string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List managed resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := app.New(opts.CompositionDir)
			if err != nil {
				return err
			}
			defer a.Close()
			filter := map[string]string{}
			if providerName != "" {
				filter["provider"] = providerName
			}
			resources, err := a.Engine.List(filter)
			if err != nil {
				return err
			}
			if len(resources) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no resources")
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), "ID\tNAME\tPROVIDER\tTYPE\tSTATUS\tREGION")
			for _, r := range resources {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\t%s\t%s\n", r.ID, r.Name, r.Provider, r.Type, r.Status, r.Region)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&providerName, "provider", "", "filter by provider")
	return cmd
}

func newStatusCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "status <id>",
		Short: "Show detailed resource status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := app.New(opts.CompositionDir)
			if err != nil {
				return err
			}
			defer a.Close()
			resource, err := a.Engine.Status(args[0])
			if err != nil {
				return err
			}
			encoded, err := json.MarshalIndent(resource, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(encoded))
			return nil
		},
	}
}

func newDestroyCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "destroy <id>",
		Short: "Destroy a managed resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := app.New(opts.CompositionDir)
			if err != nil {
				return err
			}
			defer a.Close()
			if err := a.Engine.Destroy(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "destroyed %s\n", args[0])
			return nil
		},
	}
}

func newCleanupCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Manage automated cleanup",
	}
	cmd.AddCommand(cleanupModeCommand(opts, "enable", domain.CleanupEnabled))
	cmd.AddCommand(cleanupModeCommand(opts, "disable", domain.CleanupDisabled))
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show cleanup status",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := app.New(opts.CompositionDir)
			if err != nil {
				return err
			}
			defer a.Close()
			mode, err := a.Store.CleanupMode()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), mode)
			return nil
		},
	})
	return cmd
}

func cleanupModeCommand(opts Options, use string, mode domain.CleanupMode) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: "Set cleanup " + string(mode),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := app.New(opts.CompositionDir)
			if err != nil {
				return err
			}
			defer a.Close()
			if err := a.Store.SetCleanupMode(mode); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), mode)
			return nil
		},
	}
}

func newEncryptCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "encrypt <string>",
		Short: "Encrypt a string with the master key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := app.New(opts.CompositionDir)
			if err != nil {
				return err
			}
			defer a.Close()
			ciphertext, err := a.Encryption.Encrypt(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), ciphertext)
			return nil
		},
	}
}

func newDecryptCommand(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "decrypt <string>",
		Short: "Decrypt a string with the master key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := app.New(opts.CompositionDir)
			if err != nil {
				return err
			}
			defer a.Close()
			plaintext, err := a.Encryption.Decrypt(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), plaintext)
			return nil
		},
	}
}

func parseLabels(values []string) (map[string]string, error) {
	labels := map[string]string{}
	for _, value := range values {
		key, val, ok := strings.Cut(value, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid label %q, expected key=value", value)
		}
		labels[key] = val
	}
	return labels, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
