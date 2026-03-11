package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/pagination"
	"github.com/spf13/cobra"

	"github.com/rackerlabs/goclouddns"
	"github.com/rackerlabs/goclouddns/domains"
	"github.com/rackerlabs/goclouddns/records"
	"github.com/rackerlabs/goraxauth"
)

// Build information. Populated at build-time via ldflags.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

type cliApp struct {
	timeout uint
	format  string
	wide    bool
	debug   bool
}

func main() {
	rootCmd := newRootCmd()
	rootCmd.SetArgs(normalizeLegacyArgs(os.Args[1:]))
	cobra.CheckErr(rootCmd.Execute())
}

func normalizeLegacyArgs(args []string) []string {
	if len(args) < 3 || args[0] != "record" {
		return args
	}

	switch args[2] {
	case "create", "list", "show", "update", "delete":
		rewritten := []string{"record", args[2], args[1]}
		return append(rewritten, args[3:]...)
	default:
		return args
	}
}

func newRootCmd() *cobra.Command {
	app := &cliApp{}

	rootCmd := &cobra.Command{
		Use:           "clouddns",
		Short:         "Manage Rackspace Cloud DNS domains and records",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
		Version: version,
	}

	rootCmd.SetVersionTemplate("clouddns version {{.Version}}\ncommit: " + commit + "\nbuilt: " + date + "\n")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().UintVar(&app.timeout, "timeout", 60, "Operation timeout")
	rootCmd.PersistentFlags().StringVar(&app.format, "format", "table", "output format: table or json")
	rootCmd.PersistentFlags().BoolVar(&app.wide, "wide", false, "show full-width table output")
	rootCmd.PersistentFlags().BoolVar(&app.debug, "debug", false, "show debug logging")

	rootCmd.AddCommand(newDomainCmd(app))
	rootCmd.AddCommand(newRecordCmd(app))

	return rootCmd
}

func (app *cliApp) validateOutputFormat() error {
	switch app.format {
	case "table", "json":
		return nil
	default:
		return fmt.Errorf("unsupported --format %q: must be one of table, json", app.format)
	}
}

func (app *cliApp) withService(run func(context.Context, *gophercloud.ServiceClient) error) error {
	if err := app.validateOutputFormat(); err != nil {
		return err
	}

	if app.debug {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(io.Discard)
	}

	timeout := time.Duration(app.timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	opts, err := goraxauth.AuthOptionsFromEnv()
	if err != nil {
		return err
	}

	provider, err := goraxauth.AuthenticatedClient(ctx, opts)
	if err != nil {
		return err
	}

	service, err := goclouddns.NewCloudDNS(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return err
	}

	return run(ctx, service)
}

func exactArgsValidator(n int, usage string, expected string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		switch {
		case len(args) < n:
			label := "arguments"
			if n == 1 {
				label = "argument"
			}
			return friendlyUsageError(cmd, fmt.Sprintf("missing required %s: %s", label, expected), usage)
		case len(args) > n:
			return friendlyUsageError(cmd, fmt.Sprintf("too many arguments: got %d, expected %d", len(args), n), usage)
		default:
			return nil
		}
	}
}

func noArgsValidator(usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return friendlyUsageError(cmd, "this command does not accept positional arguments", usage)
		}
		return nil
	}
}

func friendlyUsageError(cmd *cobra.Command, detail string, usage string) error {
	var b strings.Builder
	b.WriteString(detail)
	b.WriteString("\n\nUsage:\n  ")
	b.WriteString(usage)

	if cmd != nil {
		if flags := strings.TrimSpace(cmd.NonInheritedFlags().FlagUsagesWrapped(80)); flags != "" {
			b.WriteString("\n\nFlags:\n")
			b.WriteString(flags)
		}

		if flags := strings.TrimSpace(cmd.InheritedFlags().FlagUsagesWrapped(80)); flags != "" {
			b.WriteString("\n\nGlobal Flags:\n")
			b.WriteString(flags)
		}
	}

	if cmd != nil && cmd.Example != "" {
		b.WriteString("\n\nExamples:\n")
		b.WriteString(cmd.Example)
	}

	return fmt.Errorf("%s", b.String())
}

func requireAnyFlag(cmd *cobra.Command, names ...string) error {
	for _, name := range names {
		if cmd.Flags().Changed(name) {
			return nil
		}
	}

	formatted := make([]string, 0, len(names))
	for _, name := range names {
		formatted = append(formatted, "--"+name)
	}

	return fmt.Errorf("specify at least one of %s", strings.Join(formatted, ", "))
}

func printJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
}

func compactTime(ts string) string {
	if ts == "" {
		return ""
	}

	parsed, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			return ts
		}
	}

	return parsed.Format("2006-01-02 15:04")
}

func truncate(s string, max int) string {
	if len(s) <= max || max < 4 {
		return s
	}
	return s[:max-3] + "..."
}

func printDomainList(format string, wide bool, domain *domains.DomainList) error {
	if format == "json" {
		return printJSON(domain)
	}

	w := newTabWriter()
	if wide {
		fmt.Fprintln(w, "ID\tNAME\tEMAIL\tCREATED\tUPDATED")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", domain.ID, domain.Name, domain.Email, domain.Created, domain.Updated)
	} else {
		fmt.Fprintln(w, "ID\tNAME\tUPDATED")
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			domain.ID,
			truncate(domain.Name, 40),
			compactTime(domain.Updated),
		)
	}
	return w.Flush()
}

func printDomainLists(format string, wide bool, domainList []domains.DomainList) error {
	if format == "json" {
		return printJSON(domainList)
	}

	w := newTabWriter()
	if wide {
		fmt.Fprintln(w, "ID\tNAME\tEMAIL\tCREATED\tUPDATED")
		for _, domain := range domainList {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", domain.ID, domain.Name, domain.Email, domain.Created, domain.Updated)
		}
	} else {
		fmt.Fprintln(w, "ID\tNAME\tUPDATED")
		for _, domain := range domainList {
			fmt.Fprintf(w, "%s\t%s\t%s\n",
				domain.ID,
				truncate(domain.Name, 40),
				compactTime(domain.Updated),
			)
		}
	}
	return w.Flush()
}

func printDomainShow(format string, domain *domains.DomainShow) error {
	if format == "json" {
		return printJSON(domain)
	}

	nameservers := make([]string, 0, len(domain.Nameservers))
	for _, ns := range domain.Nameservers {
		nameservers = append(nameservers, ns.Name)
	}

	w := newTabWriter()
	fmt.Fprintln(w, "FIELD\tVALUE")
	fmt.Fprintf(w, "ID\t%s\n", domain.ID)
	fmt.Fprintf(w, "Name\t%s\n", domain.Name)
	fmt.Fprintf(w, "Email\t%s\n", domain.EmailAddress)
	fmt.Fprintf(w, "TTL\t%d\n", domain.TTL)
	fmt.Fprintf(w, "Comment\t%s\n", domain.Comment)
	fmt.Fprintf(w, "Account ID\t%s\n", domain.AccountID)
	fmt.Fprintf(w, "Created\t%s\n", domain.Created)
	fmt.Fprintf(w, "Updated\t%s\n", domain.Updated)
	fmt.Fprintf(w, "Nameservers\t%s\n", strings.Join(nameservers, ", "))
	fmt.Fprintf(w, "Record Count\t%d\n", domain.RecordsList.TotalEntries)
	return w.Flush()
}

func printRecordList(format string, wide bool, record *records.RecordList) error {
	if format == "json" {
		return printJSON(record)
	}

	w := newTabWriter()
	if wide {
		fmt.Fprintln(w, "ID\tNAME\tTYPE\tDATA\tTTL\tPRIORITY\tCOMMENT")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%s\n", record.ID, record.Name, record.Type, record.Data, record.TTL, record.Priority, record.Comment)
	} else {
		fmt.Fprintln(w, "ID\tNAME\tTYPE\tDATA\tTTL")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n",
			record.ID,
			truncate(record.Name, 38),
			record.Type,
			truncate(record.Data, 28),
			record.TTL,
		)
	}
	return w.Flush()
}

func printRecordLists(format string, wide bool, recordList []records.RecordList) error {
	if format == "json" {
		return printJSON(recordList)
	}

	w := newTabWriter()
	if wide {
		fmt.Fprintln(w, "ID\tNAME\tTYPE\tDATA\tTTL\tPRIORITY\tCOMMENT")
		for _, record := range recordList {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%s\n", record.ID, record.Name, record.Type, record.Data, record.TTL, record.Priority, record.Comment)
		}
	} else {
		fmt.Fprintln(w, "ID\tNAME\tTYPE\tDATA\tTTL")
		for _, record := range recordList {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n",
				record.ID,
				truncate(record.Name, 38),
				record.Type,
				truncate(record.Data, 28),
				record.TTL,
			)
		}
	}
	return w.Flush()
}

func printRecordShow(format string, record *records.RecordShow) error {
	if format == "json" {
		return printJSON(record)
	}

	w := newTabWriter()
	fmt.Fprintln(w, "FIELD\tVALUE")
	fmt.Fprintf(w, "ID\t%s\n", record.ID)
	fmt.Fprintf(w, "Name\t%s\n", record.Name)
	fmt.Fprintf(w, "Type\t%s\n", record.Type)
	fmt.Fprintf(w, "Data\t%s\n", record.Data)
	fmt.Fprintf(w, "TTL\t%d\n", record.TTL)
	fmt.Fprintf(w, "Priority\t%d\n", record.Priority)
	fmt.Fprintf(w, "Comment\t%s\n", record.Comment)
	fmt.Fprintf(w, "Created\t%s\n", record.Created)
	fmt.Fprintf(w, "Updated\t%s\n", record.Updated)
	return w.Flush()
}

func newDomainCmd(app *cliApp) *cobra.Command {
	domainCmd := &cobra.Command{
		Use:   "domain",
		Short: "Manage domains",
	}

	var createComment string
	var createTTL uint
	createCmd := &cobra.Command{
		Use:   "create DOMAIN EMAIL",
		Short: "Create a domain",
		Args:  exactArgsValidator(2, "clouddns domain create DOMAIN EMAIL", "DOMAIN and EMAIL"),
		Example: strings.Join([]string{
			"  clouddns domain create example.com admin@example.com",
			"  clouddns domain create example.com admin@example.com --ttl 7200 --comment \"production zone\"",
		}, "\n"),
		RunE: func(_ *cobra.Command, args []string) error {
			return app.withService(func(ctx context.Context, service *gophercloud.ServiceClient) error {
				opts := domains.CreateOpts{
					Name:    args[0],
					Email:   args[1],
					TTL:     createTTL,
					Comment: createComment,
				}

				domain, err := domains.Create(ctx, service, opts).Extract()
				if err != nil {
					return err
				}

				return printDomainList(app.format, app.wide, domain)
			})
		},
	}
	createCmd.Flags().StringVar(&createComment, "comment", "", "optional comments")
	createCmd.Flags().UintVar(&createTTL, "ttl", 3600, "TTL for the SOA record")

	var listName string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List domains",
		Args:  noArgsValidator("clouddns domain list"),
		Example: strings.Join([]string{
			"  clouddns domain list",
			"  clouddns domain list --name example.com",
			"  clouddns domain list --format json",
			"  clouddns domain list --wide",
		}, "\n"),
		RunE: func(_ *cobra.Command, _ []string) error {
			return app.withService(func(ctx context.Context, service *gophercloud.ServiceClient) error {
				opts := domains.ListOpts{Name: listName}
				pager := domains.List(ctx, service, opts)
				var domainList []domains.DomainList

				if err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
					pageDomains, err := domains.ExtractDomains(page)
					if err != nil {
						return false, err
					}

					domainList = append(domainList, pageDomains...)
					return true, nil
				}); err != nil {
					return err
				}

				return printDomainLists(app.format, app.wide, domainList)
			})
		},
	}
	listCmd.Flags().StringVar(&listName, "name", "", "filter domains matching this")

	showCmd := &cobra.Command{
		Use:   "show ID",
		Short: "Show a domain",
		Args:  exactArgsValidator(1, "clouddns domain show ID", "ID"),
		Example: strings.Join([]string{
			"  clouddns domain show <domain-id>",
			"  clouddns domain show <domain-id> --format json",
		}, "\n"),
		RunE: func(_ *cobra.Command, args []string) error {
			return app.withService(func(ctx context.Context, service *gophercloud.ServiceClient) error {
				domain, err := domains.Get(ctx, service, args[0]).Extract()
				if err != nil {
					return err
				}

				return printDomainShow(app.format, domain)
			})
		},
	}

	var updateEmail string
	var updateComment string
	var updateTTL uint
	updateCmd := &cobra.Command{
		Use:   "update ID",
		Short: "Update a domain",
		Args:  exactArgsValidator(1, "clouddns domain update ID", "ID"),
		Example: strings.Join([]string{
			"  clouddns domain update <domain-id> --email admin@example.com",
			"  clouddns domain update <domain-id> --ttl 7200 --comment \"updated zone\"",
		}, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAnyFlag(cmd, "email", "ttl", "comment"); err != nil {
				return err
			}

			return app.withService(func(ctx context.Context, service *gophercloud.ServiceClient) error {
				domain, err := domains.Get(ctx, service, args[0]).Extract()
				if err != nil {
					return err
				}

				opts := domains.UpdateOpts{
					Email:   updateEmail,
					TTL:     updateTTL,
					Comment: updateComment,
				}

				if err := domains.Update(ctx, service, domain, opts).ExtractErr(); err != nil {
					return err
				}

				fmt.Println("domain updated")
				return nil
			})
		},
	}
	updateCmd.Flags().StringVar(&updateEmail, "email", "", "optional email update")
	updateCmd.Flags().StringVar(&updateComment, "comment", "", "optional comments")
	updateCmd.Flags().UintVar(&updateTTL, "ttl", 0, "optional change to TTL for the SOA record")

	deleteCmd := &cobra.Command{
		Use:     "delete ID",
		Short:   "Delete a domain",
		Args:    exactArgsValidator(1, "clouddns domain delete ID", "ID"),
		Example: "  clouddns domain delete <domain-id>",
		RunE: func(_ *cobra.Command, args []string) error {
			return app.withService(func(ctx context.Context, service *gophercloud.ServiceClient) error {
				if err := domains.Delete(ctx, service, args[0]).ExtractErr(); err != nil {
					return err
				}

				fmt.Println("Successfully deleted")
				return nil
			})
		},
	}

	domainCmd.AddCommand(createCmd, listCmd, showCmd, updateCmd, deleteCmd)
	return domainCmd
}

func newRecordCmd(app *cliApp) *cobra.Command {
	recordCmd := &cobra.Command{
		Use:   "record",
		Short: "Manage records",
	}

	var createComment string
	var createTTL uint
	createCmd := &cobra.Command{
		Use:   "create DOMID NAME TYPE DATA",
		Short: "Create a record",
		Args:  exactArgsValidator(4, "clouddns record create DOMID NAME TYPE DATA", "DOMID, NAME, TYPE, and DATA"),
		Example: strings.Join([]string{
			"  clouddns record create <domain-id> app.prod.example.com A 10.5.19.11",
			"  clouddns record create <domain-id> mail.prod.example.com MX mail.example.com --ttl 300 --comment \"mail route\"",
		}, "\n"),
		RunE: func(_ *cobra.Command, args []string) error {
			return app.withService(func(ctx context.Context, service *gophercloud.ServiceClient) error {
				opts := records.CreateOpts{
					Name:    args[1],
					Type:    args[2],
					Data:    args[3],
					TTL:     createTTL,
					Comment: createComment,
				}

				record, err := records.Create(ctx, service, args[0], opts).Extract()
				if err != nil {
					return err
				}

				return printRecordList(app.format, app.wide, record)
			})
		},
	}
	createCmd.Flags().StringVar(&createComment, "comment", "", "optional comments")
	createCmd.Flags().UintVar(&createTTL, "ttl", 0, "TTL for the record")

	var listType string
	listCmd := &cobra.Command{
		Use:   "list DOMID",
		Short: "List records",
		Args:  exactArgsValidator(1, "clouddns record list DOMID", "DOMID"),
		Example: strings.Join([]string{
			"  clouddns record list <domain-id>",
			"  clouddns record list <domain-id> --type A",
			"  clouddns record list <domain-id> --format json",
			"  clouddns record list <domain-id> --wide",
		}, "\n"),
		RunE: func(_ *cobra.Command, args []string) error {
			return app.withService(func(ctx context.Context, service *gophercloud.ServiceClient) error {
				opts := records.ListOpts{
					Type: listType,
				}

				pager := records.List(ctx, service, args[0], opts)
				var recordList []records.RecordList

				if err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
					pageRecords, err := records.ExtractRecords(page)
					if err != nil {
						return false, err
					}

					recordList = append(recordList, pageRecords...)
					return true, nil
				}); err != nil {
					return err
				}

				return printRecordLists(app.format, app.wide, recordList)
			})
		},
	}
	listCmd.Flags().StringVar(&listType, "type", "", "filter records matching this type")

	showCmd := &cobra.Command{
		Use:   "show DOMID ID",
		Short: "Show a record",
		Args:  exactArgsValidator(2, "clouddns record show DOMID ID", "DOMID and ID"),
		Example: strings.Join([]string{
			"  clouddns record show <domain-id> <record-id>",
			"  clouddns record show <domain-id> <record-id> --format json",
		}, "\n"),
		RunE: func(_ *cobra.Command, args []string) error {
			return app.withService(func(ctx context.Context, service *gophercloud.ServiceClient) error {
				record, err := records.Get(ctx, service, args[0], args[1]).Extract()
				if err != nil {
					return err
				}

				return printRecordShow(app.format, record)
			})
		},
	}

	var updateData string
	var updateComment string
	var updatePriority uint
	var updateTTL uint
	updateCmd := &cobra.Command{
		Use:   "update DOMID ID",
		Short: "Update a record",
		Args:  exactArgsValidator(2, "clouddns record update DOMID ID", "DOMID and ID"),
		Example: strings.Join([]string{
			"  clouddns record update <domain-id> <record-id> --data 10.5.19.11",
			"  clouddns record <domain-id> update <record-id> --data 10.5.19.11",
		}, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAnyFlag(cmd, "data", "ttl", "priority", "comment"); err != nil {
				return err
			}

			return app.withService(func(ctx context.Context, service *gophercloud.ServiceClient) error {
				record, err := records.Get(ctx, service, args[0], args[1]).Extract()
				if err != nil {
					return err
				}

				opts := records.UpdateOpts{
					Name:     record.Name,
					Data:     updateData,
					Priority: updatePriority,
					TTL:      updateTTL,
					Comment:  updateComment,
				}

				if err := records.Update(ctx, service, args[0], record, opts).ExtractErr(); err != nil {
					return err
				}

				fmt.Println("record updated")
				return nil
			})
		},
	}
	updateCmd.Flags().StringVar(&updateData, "data", "", "optional change to data for the record")
	updateCmd.Flags().UintVar(&updatePriority, "priority", 0, "optional change to priority for the record")
	updateCmd.Flags().UintVar(&updateTTL, "ttl", 0, "optional change to TTL for the record")
	updateCmd.Flags().StringVar(&updateComment, "comment", "", "optional comments")

	deleteCmd := &cobra.Command{
		Use:     "delete DOMID ID",
		Short:   "Delete a record",
		Args:    exactArgsValidator(2, "clouddns record delete DOMID ID", "DOMID and ID"),
		Example: "  clouddns record delete <domain-id> <record-id>",
		RunE: func(_ *cobra.Command, args []string) error {
			return app.withService(func(ctx context.Context, service *gophercloud.ServiceClient) error {
				if err := records.Delete(ctx, service, args[0], args[1]).ExtractErr(); err != nil {
					return err
				}

				fmt.Println("Successfully deleted")
				return nil
			})
		},
	}

	recordCmd.AddCommand(createCmd, listCmd, showCmd, updateCmd, deleteCmd)
	return recordCmd
}
