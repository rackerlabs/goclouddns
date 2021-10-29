package domains

import (
	"encoding/json"
	"strconv"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

// DNS team changed the API from int64 to string for
// the accountId and ID fields
type StringInt int64

func (st *StringInt) UnmarshalJSON(b []byte) error {
	//convert the bytes into an interface
	//this will help us check the type of our value
	//if it is a string that can be converted into an int we convert it
	///otherwise we return an error
	var item interface{}
	if err := json.Unmarshal(b, &item); err != nil {
		return err
	}
	switch v := item.(type) {
	case int64:
		*st = StringInt(v)
	case int:
		*st = StringInt(int64(v))
	case string:
		///here convert the string into
		///an integer
		i, err := strconv.Atoi(v)
		if err != nil {
			///the string might not be of integer type
			///so return an error
			return err

		}
		*st = StringInt(int64(i))

	}
	return nil
}

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
	ID StringInt

	// Created is the date when the domain was created.
	Created string

	// Updated is the date when the domain was updated.
	Updated string

	// EmailAddress is the email associated with this domain.
	Email string `json:"emailAddress"`

	// AccountID is the Tenant ID this domain is under
	AccountID StringInt `json:"accountId"`

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
	AccountID    StringInt `json:"accountId"`
	ID           StringInt `json:"id"`
	Name         string    `json:"name"`
	EmailAddress string    `json:"emailAddress"`
	Updated      string    `json:"updated"`
	Created      string    `json:"created"`
	Comment      string    `json:"comment"`
}
