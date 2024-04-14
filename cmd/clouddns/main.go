package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/pagination"

	"github.com/rackerlabs/goclouddns"
	"github.com/rackerlabs/goclouddns/domains"
	"github.com/rackerlabs/goclouddns/records"
)

func main() {
	// create subcommand of domain
	createDomCmd := flag.NewFlagSet("create", flag.ExitOnError)
	createDomComment := createDomCmd.String("comment", "", "optional comments")
	createDomTTL := createDomCmd.Uint("ttl", 3600, "TTL for the SOA record")
	// list subcommand of domain
	listDomCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listDomFilter := listDomCmd.String("name", "", "filter domains matching this")
	// show subcommand of domain
	showDomCmd := flag.NewFlagSet("show", flag.ExitOnError)
	// update subcommand of domain
	updateDomCmd := flag.NewFlagSet("update", flag.ExitOnError)
	updateDomEmail := updateDomCmd.String("email", "", "optional email update")
	updateDomComment := updateDomCmd.String("comment", "", "optional comments")
	updateDomTTL := updateDomCmd.Uint("ttl", 0, "optional change to TTL for the SOA record")
	// delete subcommand of domain
	deleteDomCmd := flag.NewFlagSet("delete", flag.ExitOnError)

	// create subcommand of record
	createRecCmd := flag.NewFlagSet("create", flag.ExitOnError)
	createRecComment := createRecCmd.String("comment", "", "optional comments")
	createRecTTL := createRecCmd.Uint("ttl", 0, "TTL for the record")
	// list subcommand of record
	listRecCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listRecFilterName := listRecCmd.String("name", "", "filter records matching this name")
	listRecFilterData := listRecCmd.String("data", "", "filter records matching this data")
	listRecFilterType := listRecCmd.String("type", "", "filter records matching this type")
	// show subcommand of record
	showRecCmd := flag.NewFlagSet("show", flag.ExitOnError)
	// update subcommand of record
	updateRecCmd := flag.NewFlagSet("update", flag.ExitOnError)
	updateRecData := updateRecCmd.String("data", "", "optional change to data for the record")
	updateRecPriority := updateRecCmd.Uint("priority", 0, "optional change to priority for the record")
	updateRecTTL := updateRecCmd.Uint("ttl", 0, "optional change to TTL for the record")
	updateRecComment := updateRecCmd.String("comment", "", "optional comments")
	// delete subcommand of record
	deleteRecCmd := flag.NewFlagSet("delete", flag.ExitOnError)

	if len(os.Args) < 2 {
		log.Fatal("domain or record subcommand is required")
	}

	switch os.Args[1] {
	case "domain":
		if len(os.Args) < 3 {
			log.Fatal("create, list, show, update, or delete action is required")
		}

		switch os.Args[2] {
		case "create":
			createDomCmd.Parse(os.Args[3:])
		case "list":
			listDomCmd.Parse(os.Args[3:])
		case "show":
			showDomCmd.Parse(os.Args[3:])
		case "update":
			updateDomCmd.Parse(os.Args[3:])
		case "delete":
			deleteDomCmd.Parse(os.Args[3:])
		default:
			log.Fatalf("Usage: %s domain create|list|show|update|delete ...", os.Args[0])
		}
	case "record":
		if len(os.Args) < 4 {
			log.Fatal("domID and one of create, list, show, update, or delete action is required")
		}

		switch os.Args[3] {
		case "create":
			createRecCmd.Parse(os.Args[4:])
		case "list":
			listRecCmd.Parse(os.Args[4:])
		case "show":
			showRecCmd.Parse(os.Args[4:])
		case "update":
			updateRecCmd.Parse(os.Args[4:])
		case "delete":
			deleteRecCmd.Parse(os.Args[4:])
		default:
			log.Fatalf("Usage: %s record DOMID create|list|show|update|delete ...", os.Args[0])
		}
	default:
		log.Fatalf("Usage: %s domain|record ...", os.Args[0])
	}

	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		log.Fatal(err)
	}

	service, err := goclouddns.NewCloudDNS(provider, gophercloud.EndpointOpts{})
	if err != nil {
		log.Fatal(err)
	}

	if createDomCmd.Parsed() {
		args := createDomCmd.Args()
		if len(args) < 1 {
			fmt.Fprintf(createDomCmd.Output(), "Usage: %s domain create [-ttl TTL] [-comment COMMENT] DOMAIN EMAIL\n", os.Args[0])
			createDomCmd.PrintDefaults()
			os.Exit(2)
		}

		opts := domains.CreateOpts{
			Name:    args[0],
			Email:   args[1],
			TTL:     *createDomTTL,
			Comment: *createDomComment,
		}

		domain, err := domains.Create(service, opts).Extract()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", domain)

	} else if listDomCmd.Parsed() {
		opts := domains.ListOpts{
			Name: *listDomFilter,
		}

		pager := domains.List(service, opts)

		listErr := pager.EachPage(func(page pagination.Page) (bool, error) {
			domainList, err := domains.ExtractDomains(page)

			if err != nil {
				return false, err
			}

			for _, domain := range domainList {
				fmt.Printf("%+v\n", domain)
			}
			return true, err
		})

		if listErr != nil {
			log.Fatal(listErr)
		}
	} else if showDomCmd.Parsed() {
		args := showDomCmd.Args()
		if len(args) < 1 {
			fmt.Fprintf(showDomCmd.Output(), "Usage: %s domain show ID\n", os.Args[0])
			showDomCmd.PrintDefaults()
			os.Exit(2)
		}

		domID := args[0]

		domain, err := domains.Get(service, domID).Extract()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", domain)

	} else if updateDomCmd.Parsed() {
		args := updateDomCmd.Args()
		if len(args) < 1 {
			fmt.Fprintf(updateDomCmd.Output(), "Usage: %s domain update [-email EMAIL] [-ttl TTL] [-comment COMMENT] ID\n", os.Args[0])
			updateDomCmd.PrintDefaults()
			os.Exit(2)
		}

		domID := args[0]

		domain, err := domains.Get(service, domID).Extract()
		if err != nil {
			log.Fatal(err)
		}

		opts := domains.UpdateOpts{
			Email:   *updateDomEmail,
			TTL:     *updateDomTTL,
			Comment: *updateDomComment,
		}

		err = domains.Update(service, domain, opts).ExtractErr()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("domain updated")

	} else if deleteDomCmd.Parsed() {
		args := deleteDomCmd.Args()
		if len(args) < 1 {
			fmt.Fprintf(deleteDomCmd.Output(), "Usage: %s domain delete ID\n", os.Args[0])
			deleteDomCmd.PrintDefaults()
			os.Exit(2)
		}

		domID := args[0]

		deleteErr := domains.Delete(service, domID).ExtractErr()
		if deleteErr != nil {
			log.Fatal(deleteErr)
		}
		log.Print("Successfully deleted")
	} else if createRecCmd.Parsed() {
		args := createRecCmd.Args()
		if len(args) < 3 {
			fmt.Fprintf(createRecCmd.Output(), "Usage: %s record DOMID create [-ttl TTL] [-comment COMMENT] NAME TYPE DATA\n", os.Args[0])
			createRecCmd.PrintDefaults()
			os.Exit(2)
		}

		domID := os.Args[2]

		opts := records.CreateOpts{
			Name:    args[0],
			Type:    args[1],
			Data:    args[2],
			TTL:     *createRecTTL,
			Comment: *createRecComment,
		}

		record, err := records.Create(service, domID, opts).Extract()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", record)
	} else if listRecCmd.Parsed() {
		domID := os.Args[2]

		opts := records.ListOpts{
			Name: *listRecFilterName,
			Data: *listRecFilterData,
			Type: *listRecFilterType,
		}

		// set the default to A records if one wasn't supplied
		if *listRecFilterName != "" && *listRecFilterData != "" && *listRecFilterType == "" {
			opts.Type = "A"
		}

		pager := records.List(service, domID, opts)

		listErr := pager.EachPage(func(page pagination.Page) (bool, error) {
			recordList, err := records.ExtractRecords(page)

			if err != nil {
				return false, err
			}

			for _, record := range recordList {
				fmt.Printf("%+v\n", record)
			}
			return true, err
		})

		if listErr != nil {
			log.Fatal(listErr)
		}
	} else if showRecCmd.Parsed() {
		args := showRecCmd.Args()
		if len(args) < 1 {
			fmt.Fprintf(showRecCmd.Output(), "Usage: %s record DOMID show ID\n", os.Args[0])
			showRecCmd.PrintDefaults()
			os.Exit(2)
		}

		domID := os.Args[2]

		record, err := records.Get(service, domID, args[0]).Extract()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", record)
	} else if updateRecCmd.Parsed() {
		args := updateRecCmd.Args()
		if len(args) < 1 {
			fmt.Fprintf(updateRecCmd.Output(), "Usage: %s record DOMID update [-data DATA] [-ttl TTL] [-priority PRIORITY] ID\n", os.Args[0])
			updateRecCmd.PrintDefaults()
			os.Exit(2)
		}

		domID := os.Args[2]

		record, err := records.Get(service, domID, args[0]).Extract()
		if err != nil {
			log.Fatal(err)
		}

		opts := records.UpdateOpts{
			Name:     record.Name,
			Data:     *updateRecData,
			Priority: *updateRecPriority,
			TTL:      *updateRecTTL,
			Comment:  *updateRecComment,
		}

		err = records.Update(service, domID, record, opts).ExtractErr()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("record updated")
	} else if deleteRecCmd.Parsed() {
		args := deleteRecCmd.Args()
		if len(args) < 1 {
			fmt.Fprintf(deleteRecCmd.Output(), "Usage: %s record DOMID delete ID\n", os.Args[0])
			deleteRecCmd.PrintDefaults()
			os.Exit(2)
		}

		domID := os.Args[2]

		deleteErr := records.Delete(service, domID, args[0]).ExtractErr()
		if deleteErr != nil {
			log.Fatal(deleteErr)
		}
		log.Print("Successfully deleted")
	} else {
		flag.PrintDefaults()
		os.Exit(1)
	}
}
