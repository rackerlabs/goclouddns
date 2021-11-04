package domains

import (
	"log"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"

	"github.rackspace.com/doug1840/goclouddns"
)

// ListOptsBuilder allows extensions to add additional parameters to the
// List request.
type ListOptsBuilder interface {
	ToDomainListQuery() (string, error)
}

// ListOpts contain options filtering Domains returned from a call to List.
type ListOpts struct {
	// Name is the name of the Domain.
	Name string `q:"name"`
}

// ToDomainListQuery formats a ListOpts into a query string.
func (opts ListOpts) ToDomainListQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	return q.String(), err
}

func List(client *gophercloud.ServiceClient, opts ListOptsBuilder) pagination.Pager {
	url := client.ServiceURL("domains")
	if opts != nil {
		query, err := opts.ToDomainListQuery()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query
	}

	log.Printf("GET %s", url)

	return pagination.NewPager(client, url, func(r pagination.PageResult) pagination.Page {
		return DomainPage{pagination.LinkedPageBase{PageResult: r}}
	})
}

// Get returns data about a specific domain by its ID.
func Get(client *gophercloud.ServiceClient, id string) (r GetResult) {
	url := client.ServiceURL("domains", id)
	log.Printf("GET %s", url)
	_, r.Err = client.Get(url, &r.Body, nil)
	return
}

// Delete deletes the specified domain ID.
func Delete(client *gophercloud.ServiceClient, id string) (r DeleteResult) {
	url := client.ServiceURL("domains", id)
	log.Printf("DELETE %s", url)

	var resp goclouddns.AsyncResult
	_, resp.Err = client.Delete(url, &gophercloud.RequestOpts{
		JSONResponse: &resp.Body,
	})

	if resp.Err != nil {
		r.Err = resp.Err
		return
	}

	if err := goclouddns.WaitForStatus(client, &resp, "COMPLETED"); err != nil {
		r.Err = err
		return
	}
	r.Body = resp.Body
	return
}

// CreateOpts contain the values necessary to create a domain
type CreateOpts struct {
	// Name is the name of the Domain.
	Name    string `json:"name"`
	Email   string `json:"emailAddress"`
	TTL     uint   `json:"ttl"`
	Comment string `json:"comment"`
}

// Create creates a requested domain
func Create(client *gophercloud.ServiceClient, opts CreateOpts) (r CreateResult) {
	url := client.ServiceURL("domains")

	if opts.TTL == 0 {
		opts.TTL = 3600
	}

	log.Printf("POST %s", url)

	var body = struct {
		Domains []CreateOpts `json:"domains"`
	}{
		[]CreateOpts{opts},
	}

	var resp goclouddns.AsyncResult
	_, resp.Err = client.Post(url, body, &resp.Body, nil)
	if resp.Err != nil {
		r.Err = resp.Err
		return
	}

	if err := goclouddns.WaitForStatus(client, &resp, "COMPLETED"); err != nil {
		r.Err = err
		return
	}
	r.Body = resp.Body
	return
}

// UpdateOpts contain the values necessary to create a domain
type UpdateOpts struct {
	Email   string `json:"emailAddress,omitempty"`
	TTL     uint   `json:"ttl,omitempty"`
	Comment string `json:"comment,omitempty"`
}

// Update updates a requested domain
func Update(client *gophercloud.ServiceClient, domain *DomainShow, opts UpdateOpts) (r UpdateResult) {
	url := client.ServiceURL("domains", domain.ID)

	log.Printf("PUT %s", url)

	var resp goclouddns.AsyncResult
	_, resp.Err = client.Put(url, opts, &resp.Body, nil)
	if resp.Err != nil {
		r.Err = resp.Err
		return
	}

	if err := goclouddns.WaitForStatus(client, &resp, "COMPLETED"); err != nil {
		r.Err = err
		return
	}
	r.Body = resp.Body
	return
}
