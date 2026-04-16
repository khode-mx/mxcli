// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/mendixlabs/mxcli/internal/auth"
	"github.com/mendixlabs/mxcli/internal/catalog"
	"github.com/spf13/cobra"
)

var catalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Search and manage Mendix Catalog services",
	Long: `Search for data sources and services registered in Mendix Catalog.

Requires authentication via Personal Access Token (PAT). Create a PAT at:
  https://user-settings.mendix.com/

Storage priority:
  1. MENDIX_PAT env var (set MXCLI_PROFILE to target a non-default profile)
  2. ~/.mxcli/auth.json (mode 0600)`,
}

var catalogSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for services in the Catalog",
	Long: `Search for data sources and services in Mendix Catalog.

Examples:
  mxcli catalog search "customer"
  mxcli catalog search "order" --service-type OData
  mxcli catalog search "api" --production-only --json`,
	Args: cobra.ExactArgs(1),
	RunE: runCatalogSearch,
}

func init() {
	catalogSearchCmd.Flags().String("profile", auth.ProfileDefault, "credential profile name")
	catalogSearchCmd.Flags().String("service-type", "", "filter by service type (OData, REST, SOAP)")
	catalogSearchCmd.Flags().Bool("production-only", false, "show only production endpoints")
	catalogSearchCmd.Flags().Bool("owned-only", false, "show only owned services")
	catalogSearchCmd.Flags().Int("limit", 20, "results per page (max 100)")
	catalogSearchCmd.Flags().Int("offset", 0, "pagination offset")
	catalogSearchCmd.Flags().Bool("json", false, "output as JSON array")

	catalogCmd.AddCommand(catalogSearchCmd)
	rootCmd.AddCommand(catalogCmd)
}

func runCatalogSearch(cmd *cobra.Command, args []string) error {
	query := args[0]
	profile, _ := cmd.Flags().GetString("profile")
	serviceType, _ := cmd.Flags().GetString("service-type")
	prodOnly, _ := cmd.Flags().GetBool("production-only")
	ownedOnly, _ := cmd.Flags().GetBool("owned-only")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	asJSON, _ := cmd.Flags().GetBool("json")

	// Create client
	client, err := catalog.NewClient(cmd.Context(), profile)
	if err != nil {
		if _, ok := err.(*auth.ErrNoCredential); ok {
			return fmt.Errorf("no credential found. Run: mxcli auth login")
		}
		return err
	}

	// Execute search
	opts := catalog.SearchOptions{
		Query:                   query,
		ServiceType:             serviceType,
		ProductionEndpointsOnly: prodOnly,
		OwnedContentOnly:        ownedOnly,
		Limit:                   limit,
		Offset:                  offset,
	}
	resp, err := client.Search(cmd.Context(), opts)
	if err != nil {
		if _, ok := err.(*auth.ErrUnauthenticated); ok {
			return fmt.Errorf("authentication failed. Run: mxcli auth login")
		}
		return err
	}

	// Output
	if asJSON {
		return outputJSON(cmd, resp.Data)
	}
	return outputTable(cmd, resp)
}

func outputTable(cmd *cobra.Command, resp *catalog.SearchResponse) error {
	if len(resp.Data) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No results found.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tVERSION\tAPPLICATION\tENVIRONMENT\tPROD\tUUID")

	for _, item := range resp.Data {
		name := truncate(item.Name, 22)
		typ := truncate(item.ServiceType, 8)
		version := truncate(item.Version, 10)
		app := truncate(item.Application.Name, 20)
		env := truncate(item.Environment.Type, 12)
		prod := ""
		if item.Environment.Type == "Production" {
			prod = "Yes"
		}
		uuid := item.UUID
		if len(uuid) >= 8 {
			uuid = uuid[:8] // Short UUID
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			name, typ, version, app, env, prod, uuid)
	}

	fmt.Fprintf(w, "\nTotal: %d results (showing %d-%d)\n",
		resp.TotalResults, resp.Offset+1, resp.Offset+len(resp.Data))

	return w.Flush()
}

func outputJSON(cmd *cobra.Command, data []catalog.SearchResult) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
