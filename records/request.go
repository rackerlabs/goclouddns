package records

import (
	"log"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"

	"github.com/rackerlabs/goclouddns"
)

// ListOptsBuilder allows extensions to add additional parameters to the
// List request.
type ListOptsBuilder interface {
	ToRecordListQuery() (string, error)
}

// ListOpts contain options filtering Records returned from a call to List.
type ListOpts struct {
	// Name is the name of the Record.
	Name string `q:"name"`
	Data string `q:"data"`
	Type string `q:"type"`
}

// ToRecordListQuery formats a ListOpts into a query string.
func (opts ListOpts) ToRecordListQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	return q.String(), err
}

func List(client *gophercloud.ServiceClient, domID string, opts ListOptsBuilder) pagination.Pager {
	url := client.ServiceURL("domains", domID, "records")
	if opts != nil {
		query, err := opts.ToRecordListQuery()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query
	}

	log.Printf("GET %s", url)

	return pagination.NewPager(client, url, func(r pagination.PageResult) pagination.Page {
		return RecordPage{pagination.LinkedPageBase{PageResult: r}}
	})
}

// Get returns data about a specific record by its ID.
func Get(client *gophercloud.ServiceClient, domID string, id string) (r GetResult) {
	url := client.ServiceURL("domains", domID, "records", id)
	log.Printf("GET %s", url)
	_, r.Err = client.Get(url, &r.Body, nil)
	return
}

// Delete deletes the specified record ID.
func Delete(client *gophercloud.ServiceClient, domID string, id string) (r DeleteResult) {
	url := client.ServiceURL("domains", domID, "records", id)
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

// CreateOpts contain the values necessary to create a record
type CreateOpts struct {
	// Name is the name of the Record.
	Name     string `json:"name"`
	Type     string `json:"type"`
	Data     string `json:"data"`
	TTL      uint   `json:"ttl,omitempty"`
	Comment  string `json:"comment,omitempty"`
	Priority uint   `json:"priority,omitempty"`
}

// Create creates a requested record
func Create(client *gophercloud.ServiceClient, domID string, opts CreateOpts) (r CreateResult) {
	url := client.ServiceURL("domains", domID, "records")

	log.Printf("POST %s", url)

	var body = struct {
		Records []CreateOpts `json:"records"`
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

// UpdateOpts contain the values necessary to create a record
type UpdateOpts struct {
	Name     string `json:"name"`
	Data     string `json:"data"`
	TTL      uint   `json:"ttl,omitempty"`
	Comment  string `json:"comment,omitempty"`
	Priority uint   `json:"priority,omitempty"`
}

// Update updates a requested record
func Update(client *gophercloud.ServiceClient, domID string, record *RecordShow, opts UpdateOpts) (r UpdateResult) {
	url := client.ServiceURL("domains", domID, "records", record.ID)

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
