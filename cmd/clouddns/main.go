package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/pagination"

	"github.rackspace.com/doug1840/goclouddns"
	"github.rackspace.com/doug1840/goclouddns/domains"
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
	// list subcommand of record
	listRecCmd := flag.NewFlagSet("list", flag.ExitOnError)
	// show subcommand of record
	showRecCmd := flag.NewFlagSet("show", flag.ExitOnError)
	// update subcommand of record
	updateRecCmd := flag.NewFlagSet("update", flag.ExitOnError)
	// delete subcommand of record
	deleteRecCmd := flag.NewFlagSet("delete", flag.ExitOnError)

	if len(os.Args) < 2 {
		log.Fatal("domain or record subcommand is required")
	}

	switch os.Args[1] {
	case "domain":
		if len(os.Args) < 3 {
			log.Fatal("create, show, update, or delete action is required")
		}

		switch os.Args[2] {
		case "create":
			createDomCmd.Parse(os.Args[3:])
			if len(os.Args) < 5 {
				fmt.Println(os.Args[0], "domain create [domain] [email]")
				fmt.Println("")
				createDomCmd.PrintDefaults()
				os.Exit(2)
			}
		case "list":
			listDomCmd.Parse(os.Args[3:])
		case "show":
			showDomCmd.Parse(os.Args[3:])
			if len(os.Args) < 4 {
				fmt.Println(os.Args[0], "domain show [ID]")
				os.Exit(2)
			}
		case "update":
			updateDomCmd.Parse(os.Args[3:])
			if len(os.Args) < 4 {
				fmt.Println(os.Args[0], "domain show [ID]")
				os.Exit(2)
			}
		case "delete":
			deleteDomCmd.Parse(os.Args[3:])
			if len(os.Args) < 4 {
				fmt.Println(os.Args[0], "domain show [ID]")
				os.Exit(2)
			}
		default:
			flag.PrintDefaults()
			os.Exit(1)
		}
	case "record":
		if len(os.Args) < 4 {
			log.Fatal("domID and one of create, show, update, or delete action is required")
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
			flag.PrintDefaults()
			os.Exit(1)
		}
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
		opts := domains.CreateOpts{
			Name:    os.Args[3],
			Email:   os.Args[4],
			TTL:     *createDomTTL,
			Comment: *createDomComment,
		}

		domain, err := domains.Create(service, opts).Extract()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", domain)

	} else if listDomCmd.Parsed() {
		opts := domains.ListOpts{}

		if *listDomFilter != "" {
			opts.Name = *listDomFilter
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
		domID, err := strconv.ParseUint(os.Args[3], 10, 64)
		if err != nil {
			log.Fatal("invalid domain id: %s", err)
		}

		domain, err := domains.Get(service, domID).Extract()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", domain)

	} else if updateDomCmd.Parsed() {
		domID, err := strconv.ParseUint(os.Args[3], 10, 64)
		if err != nil {
			log.Fatal("invalid domain id: %s", err)
		}

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
		domID, err := strconv.ParseUint(os.Args[3], 10, 64)
		if err != nil {
			log.Fatal("invalid domain id: %s", err)
		}

		deleteErr := domains.Delete(service, domID).ExtractErr()
		if deleteErr != nil {
			log.Fatal(deleteErr)
		}
		log.Print("Successfully deleted")
	} else if createRecCmd.Parsed() {
		fmt.Println("record create")
	} else if showRecCmd.Parsed() {
		fmt.Println("record show")
	} else if updateRecCmd.Parsed() {
		fmt.Println("record update")
	} else if deleteRecCmd.Parsed() {
		fmt.Println("record delete")
	} else {
		flag.PrintDefaults()
		os.Exit(1)
	}
}
