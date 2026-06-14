package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/atotto/clipboard"
	"github.com/khrees/veilo/models"
	"github.com/spf13/cobra"
)

type CLIConfig struct {
	APIURL        string `json:"api_url"`
	APIKey        string `json:"api_key"`
	DefaultDomain string `json:"default_domain"`
	DefaultEmail  string `json:"default_email"`
}

var jsonOutput bool

var rootCmd = &cobra.Command{
	Use:   "veilo",
	Short: "Veilo CLI - The Invisible Email Shield",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output raw JSON")

	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(enableCmd)
	rootCmd.AddCommand(disableCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(logsCmd)

	domainsCmd.AddCommand(domainsListCmd)
	domainsCmd.AddCommand(domainsAddCmd)
	domainsCmd.AddCommand(domainsRemoveCmd)
	rootCmd.AddCommand(domainsCmd)

	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)

	// Register command-specific flags
	listCmd.Flags().Bool("enabled", false, "Filter by enabled aliases only")

	createCmd.Flags().String("slug", "", "Custom slug for the alias")
	createCmd.Flags().String("domain", "", "Custom domain for the alias (fallback to config default-domain)")
	createCmd.Flags().String("email", "", "Real destination email (fallback to config default-email)")
	createCmd.Flags().String("label", "", "Optional label for the alias")
	createCmd.Flags().Bool("copy", false, "Copy created address to clipboard")
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "veilo")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func loadConfig() (*CLIConfig, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &CLIConfig{}, nil
		}
		return nil, err
	}
	defer file.Close()
	var cfg CLIConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func saveConfig(cfg *CLIConfig) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}

func makeRequest(method, path string, body any, target any) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg.APIURL == "" {
		return errors.New("API URL is not set. Run: veilo config set api-url <url>")
	}

	url := strings.TrimSuffix(cfg.APIURL, "/") + path

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		if apiErr.Message != "" {
			return fmt.Errorf("API error (%d): %s", resp.StatusCode, apiErr.Message)
		}
		return fmt.Errorf("API error (%d)", resp.StatusCode)
	}

	if target != nil {
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
			return err
		}
		if dataJSON, ok := raw["data"]; ok {
			return json.Unmarshal(dataJSON, target)
		}
	}
	return nil
}

func formatRelativeTime(t *time.Time) string {
	if t == nil {
		return "never"
	}
	diff := time.Since(*t)
	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
}

// ----------------------------------------------------
// SERVER
// ----------------------------------------------------
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Veilo API and Webhook Server",
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

// ----------------------------------------------------
// ALIASES
// ----------------------------------------------------
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all email aliases",
	RunE: func(cmd *cobra.Command, args []string) error {
		enabled, _ := cmd.Flags().GetBool("enabled")
		path := "/aliases"
		if enabled {
			path += "?enabled=true"
		}

		var aliases []models.Alias
		if err := makeRequest(http.MethodGet, path, nil, &aliases); err != nil {
			return err
		}

		if jsonOutput {
			b, _ := json.MarshalIndent(aliases, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "  ADDRESS\tLABEL\tSTATUS\tFORWARDED\tLAST USED")
		for _, a := range aliases {
			status := "disabled"
			if a.Enabled {
				status = "enabled"
			}
			label := "—"
			if a.Label != nil && *a.Label != "" {
				label = *a.Label
			}
			lastUsed := formatRelativeTime(a.LastUsedAt)
			fmt.Fprintf(w, "  %s\t%s\t%s\t%d\t%s\n", a.Address, label, status, a.ForwardCount, lastUsed)
		}
		w.Flush()
		return nil
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new email alias",
	RunE: func(cmd *cobra.Command, args []string) error {
		slug, _ := cmd.Flags().GetString("slug")
		domain, _ := cmd.Flags().GetString("domain")
		email, _ := cmd.Flags().GetString("email")
		label, _ := cmd.Flags().GetString("label")
		copyToClipboard, _ := cmd.Flags().GetBool("copy")

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		if domain == "" {
			domain = cfg.DefaultDomain
		}
		if email == "" {
			email = cfg.DefaultEmail
		}

		if domain == "" || email == "" {
			return errors.New("domain and email are required. Specify via flags (--domain, --email) or set defaults in config:\n  veilo config set default-domain <domain>\n  veilo config set default-email <email>")
		}

		body := map[string]any{
			"slug":       slug,
			"domain":     domain,
			"real_email": email,
		}
		if label != "" {
			body["label"] = label
		}

		var alias models.Alias
		if err := makeRequest(http.MethodPost, "/aliases", body, &alias); err != nil {
			return err
		}

		if jsonOutput {
			b, _ := json.MarshalIndent(alias, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		fmt.Printf("Created alias: %s\n", alias.Address)
		if copyToClipboard {
			if err := clipboard.WriteAll(alias.Address); err != nil {
				fmt.Printf("Warning: Failed to copy to clipboard: %v\n", err)
			} else {
				fmt.Println("Copied email address to clipboard.")
			}
		}
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get <address-or-id>",
	Short: "Get details of a specific alias",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var alias models.Alias
		if err := makeRequest(http.MethodGet, "/aliases/"+args[0], nil, &alias); err != nil {
			return err
		}

		if jsonOutput {
			b, _ := json.MarshalIndent(alias, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "  ID:\t%s\n", alias.ID)
		fmt.Fprintf(w, "  ADDRESS:\t%s\n", alias.Address)
		fmt.Fprintf(w, "  SLUG:\t%s\n", alias.Slug)
		fmt.Fprintf(w, "  DOMAIN:\t%s\n", alias.Domain)
		fmt.Fprintf(w, "  REAL EMAIL:\t%s\n", alias.RealEmail)
		label := "—"
		if alias.Label != nil && *alias.Label != "" {
			label = *alias.Label
		}
		fmt.Fprintf(w, "  LABEL:\t%s\n", label)
		status := "disabled"
		if alias.Enabled {
			status = "enabled"
		}
		fmt.Fprintf(w, "  STATUS:\t%s\n", status)
		fmt.Fprintf(w, "  FORWARDED:\t%d\n", alias.ForwardCount)
		lastUsed := "never"
		if alias.LastUsedAt != nil {
			lastUsed = alias.LastUsedAt.Format(time.RFC1123)
		}
		fmt.Fprintf(w, "  LAST USED:\t%s\n", lastUsed)
		fmt.Fprintf(w, "  CREATED:\t%s\n", alias.CreatedAt.Format(time.RFC1123))
		w.Flush()
		return nil
	},
}

var enableCmd = &cobra.Command{
	Use:   "enable <address-or-id>",
	Short: "Enable an alias",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{"enabled": true}
		if err := makeRequest(http.MethodPut, "/aliases/"+args[0], body, nil); err != nil {
			return err
		}
		fmt.Printf("Alias %s enabled.\n", args[0])
		return nil
	},
}

var disableCmd = &cobra.Command{
	Use:   "disable <address-or-id>",
	Short: "Disable an alias",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{"enabled": false}
		if err := makeRequest(http.MethodPut, "/aliases/"+args[0], body, nil); err != nil {
			return err
		}
		fmt.Printf("Alias %s disabled.\n", args[0])
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <address-or-id>",
	Short: "Delete an alias",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := makeRequest(http.MethodDelete, "/aliases/"+args[0], nil, nil); err != nil {
			return err
		}
		fmt.Printf("Alias %s deleted.\n", args[0])
		return nil
	},
}

// ----------------------------------------------------
// STATS
// ----------------------------------------------------
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "View aggregate statistics",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var stats models.Stats
		if err := makeRequest(http.MethodGet, "/stats", nil, &stats); err != nil {
			return err
		}

		if jsonOutput {
			b, _ := json.MarshalIndent(stats, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "  TOTAL ALIASES:\t%d\n", stats.TotalAliases)
		fmt.Fprintf(w, "  TOTAL FORWARDED:\t%d\n", stats.TotalForwarded)
		fmt.Fprintf(w, "  TOTAL BLOCKED:\t%d\n", stats.TotalBlocked)
		w.Flush()
		return nil
	},
}

// ----------------------------------------------------
// LOGS
// ----------------------------------------------------
var logsCmd = &cobra.Command{
	Use:   "logs <address-or-id>",
	Short: "Get forward logs for a specific alias",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var logs []models.ForwardLog
		if err := makeRequest(http.MethodGet, "/aliases/"+args[0]+"/logs", nil, &logs); err != nil {
			return err
		}

		if jsonOutput {
			b, _ := json.MarshalIndent(logs, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		if len(logs) == 0 {
			fmt.Println("No logs found for this alias.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "  TIMESTAMP\tDIRECTION\tSENDER\tSUBJECT\tSTATUS")
		for _, l := range logs {
			sender := "—"
			if l.Sender != nil {
				sender = *l.Sender
			}
			subject := "—"
			if l.Subject != nil {
				subject = *l.Subject
			}
			fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\n",
				l.CreatedAt.Format("2006-01-02 15:04:02"),
				l.Direction,
				sender,
				subject,
				l.Status,
			)
		}
		w.Flush()
		return nil
	},
}

// ----------------------------------------------------
// DOMAINS
// ----------------------------------------------------
var domainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "Manage domains configured in Veilo",
}

var domainsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered domains",
	RunE: func(cmd *cobra.Command, args []string) error {
		var domains []models.Domain
		if err := makeRequest(http.MethodGet, "/domains", nil, &domains); err != nil {
			return err
		}

		if jsonOutput {
			b, _ := json.MarshalIndent(domains, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "  DOMAIN\tSTATUS\tCREATED AT")
		for _, d := range domains {
			status := "unverified"
			if d.Verified {
				status = "verified"
			}
			fmt.Fprintf(w, "  %s\t%s\t%s\n", d.Name, status, d.CreatedAt.Format("2006-01-02 15:04:02"))
		}
		w.Flush()
		return nil
	},
}

var domainsAddCmd = &cobra.Command{
	Use:   "add <domain>",
	Short: "Register and configure a new domain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{"domain": args[0]}
		if err := makeRequest(http.MethodPost, "/domains", body, nil); err != nil {
			return err
		}
		fmt.Printf("Domain %s registered successfully.\n", args[0])
		return nil
	},
}

var domainsRemoveCmd = &cobra.Command{
	Use:   "remove <domain>",
	Short: "Remove a domain registration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := makeRequest(http.MethodDelete, "/domains/"+args[0], nil, nil); err != nil {
			return err
		}
		fmt.Printf("Domain %s removed successfully.\n", args[0])
		return nil
	},
}

// ----------------------------------------------------
// CONFIGURATION
// ----------------------------------------------------
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		key := strings.ToLower(args[0])
		val := args[1]

		switch key {
		case "api-url":
			cfg.APIURL = val
		case "api-key":
			cfg.APIKey = val
		case "default-domain":
			cfg.DefaultDomain = val
		case "default-email":
			cfg.DefaultEmail = val
		default:
			return fmt.Errorf("unknown config key: %s (valid options: api-url, api-key, default-domain, default-email)", args[0])
		}

		if err := saveConfig(cfg); err != nil {
			return err
		}
		fmt.Printf("Config key %s set to %s.\n", key, val)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		key := strings.ToLower(args[0])
		switch key {
		case "api-url":
			fmt.Println(cfg.APIURL)
		case "api-key":
			fmt.Println(cfg.APIKey)
		case "default-domain":
			fmt.Println(cfg.DefaultDomain)
		case "default-email":
			fmt.Println(cfg.DefaultEmail)
		default:
			return fmt.Errorf("unknown config key: %s", args[0])
		}
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show all configuration values",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "  API URL:\t%s\n", cfg.APIURL)
		fmt.Fprintf(w, "  API KEY:\t%s\n", cfg.APIKey)
		fmt.Fprintf(w, "  DEFAULT DOMAIN:\t%s\n", cfg.DefaultDomain)
		fmt.Fprintf(w, "  DEFAULT EMAIL:\t%s\n", cfg.DefaultEmail)
		w.Flush()
		return nil
	},
}
