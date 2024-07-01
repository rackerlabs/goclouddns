package records

import (
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/pagination"
)

// CreateResult is the result of a Create operation
type CreateResult struct {
	gophercloud.Result
}

// Extract interprets a CreateResult as a Record.
func (r CreateResult) Extract() (*RecordList, error) {
	var s struct {
		Response struct {
			Records []*RecordList `json:"records"`
		} `json:"response"`
	}
	err := r.ExtractInto(&s)
	if err != nil {
		return nil, err
	}
	return s.Response.Records[0], err
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
// interpret it as a Record.
type GetResult struct {
	gophercloud.Result
}

// DeleteResult is the result from a Delete operation. Call its ExtractErr

// Extract interprets a GetResult as a Record.
func (r GetResult) Extract() (*RecordShow, error) {
	var s RecordShow
	err := r.ExtractInto(&s)
	return &s, err
}

// RecordList represents a record returned by the CloudDNS API.
type RecordList struct {
	// ID is the unique ID of a record.
	ID string

	// name is the record name
	Name string

	// type is the record type
	Type string

	// data is the record data
	Data string

	// TTL of the record
	TTL uint

	// priority for SRV and MX records
	Priority uint

	// optional comment for the record
	Comment string
}

// RecordPage contains a single page of all Records returne from a List
// operation. Use ExtractRecords to convert it into a slice of usable structs.
type RecordPage struct {
	pagination.LinkedPageBase
}

// IsEmpty returns true if response contains no Record results.
func (r RecordPage) IsEmpty() (bool, error) {
	records, err := ExtractRecords(r)
	return len(records) == 0, err
}

// NextPageURL uses the response's embedded link reference to navigate to the
// next page of results.
func (page RecordPage) NextPageURL() (string, error) {
	var s struct {
		Links []gophercloud.Link `json:"links"`
	}
	err := page.ExtractInto(&s)
	if err != nil {
		return "", err
	}
	return gophercloud.ExtractNextURL(s.Links)
}

// ExtractRecords converts a page of List results into a slice of usable Record
// structs.
func ExtractRecords(r pagination.Page) ([]RecordList, error) {
	var s struct {
		Records []RecordList `json:"records"`
	}
	err := (r.(RecordPage)).ExtractInto(&s)
	return s.Records, err
}

// RecordShow represents a record returned by the CloudDNS API.
type RecordShow struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Data     string `json:"data"`
	TTL      uint   `json:"ttl"`
	Priority uint   `json:"priority"`
	Comment  string `json:"comment"`
	Updated  string `json:"updated"`
	Created  string `json:"created"`
}
