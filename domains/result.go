package domains

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

// CreateResult is the result of a Create operation
type CreateResult struct {
	gophercloud.Result
}

// Extract interprets a CreateResult as a Domain.
func (r CreateResult) Extract() (*DomainList, error) {
	var s struct {
		Response struct {
			Domains []*DomainList `json:"domains"`
		} `json:"response"`
	}
	err := r.ExtractInto(&s)
	return s.Response.Domains[0], err
}

// method to determine if the call succeeded or failed.
type DeleteResult struct {
	gophercloud.ErrResult
}

// method to determine if the call succeeded or failed.
type UpdateResult struct {
	gophercloud.ErrResult
}

// GetResult is the response from a Get operation. Call its Extract method to
// interpret it as a Domain.
type GetResult struct {
	gophercloud.Result
}

// DeleteResult is the result from a Delete operation. Call its ExtractErr

// Extract interprets a GetResult as a Domain.
func (r GetResult) Extract() (*DomainShow, error) {
	var s DomainShow
	err := r.ExtractInto(&s)
	return &s, err
}

// DomainList represents a domain returned by the CloudDNS API.
type DomainList struct {
	// ID is the unique ID of a domain.
	ID uint64

	// Created is the date when the domain was created.
	Created string

	// Updated is the date when the domain was updated.
	Updated string

	// EmailAddress is the email associated with this domain.
	Email string `json:"emailAddress"`

	// AccountID is the Tenant ID this domain is under
	AccountID uint64 `json:"accountId"`

	// name is the domain name
	Name string
}

// DomainPage contains a single page of all Domains returne from a List
// operation. Use ExtractDomains to convert it into a slice of usable structs.
type DomainPage struct {
	pagination.LinkedPageBase
}

// IsEmpty returns true if response contains no Domain results.
func (r DomainPage) IsEmpty() (bool, error) {
	domains, err := ExtractDomains(r)
	return len(domains) == 0, err
}

// NextPageURL uses the response's embedded link reference to navigate to the
// next page of results.
func (page DomainPage) NextPageURL() (string, error) {
	var s struct {
		Links []gophercloud.Link `json:"links"`
	}
	err := page.ExtractInto(&s)
	if err != nil {
		return "", err
	}
	return gophercloud.ExtractNextURL(s.Links)
}

// ExtractDomains converts a page of List results into a slice of usable Domain
// structs.
func ExtractDomains(r pagination.Page) ([]DomainList, error) {
	var s struct {
		Domains []DomainList `json:"domains"`
	}
	err := (r.(DomainPage)).ExtractInto(&s)
	return s.Domains, err
}

// DomainShow represents a domain returned by the CloudDNS API.
type DomainShow struct {
	RecordsList struct {
		TotalEntries int `json:"totalEntries"`
		Records      []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Type    string `json:"type"`
			Data    string `json:"data"`
			TTL     uint   `json:"ttl"`
			Updated string `json:"updated"`
			Created string `json:"created"`
		} `json:"records"`
	} `json:"recordsList"`
	TTL         uint64 `json:"ttl"`
	Nameservers []struct {
		Name string `json:"name"`
	} `json:"nameservers"`
	AccountID    int64  `json:"accountId"`
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	EmailAddress string `json:"emailAddress"`
	Updated      string `json:"updated"`
	Created      string `json:"created"`
	Comment      string `json:"comment"`
}
