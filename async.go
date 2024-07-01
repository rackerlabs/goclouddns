package goclouddns

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2"
)

// AsyncResult is the result of any async operation
type AsyncResult struct {
	gophercloud.Result
}

type AsyncMessage struct {
	CallbackURL string                 `json:"callbackUrl"`
	JobID       string                 `json:"jobId"`
	Request     string                 `json:"request"`
	Response    map[string]interface{} `json:"response"`
	Error       map[string]interface{} `json:"error"`
	RequestURL  string                 `json:"requestUrl"`
	Verb        string                 `json:"verb"`
	Status      string                 `json:"status"`
}

// Extract interprets a GetResult as a Domain.
func (r AsyncResult) Extract() (*AsyncMessage, error) {
	var s AsyncMessage
	err := r.ExtractInto(&s)
	return &s, err
}

func WaitForStatus(ctx context.Context, client *gophercloud.ServiceClient, ret *AsyncResult, status string) error {
	req, err := ret.Extract()
	if err != nil {
		return err
	}

	url := req.CallbackURL + "?showDetails=true"

	return gophercloud.WaitFor(ctx, func(ctx context.Context) (bool, error) {
		var resp gophercloud.Result
		if _, err := client.Get(ctx, url, &resp.Body, nil); err != nil {
			return false, err
		}

		var latest AsyncMessage
		if err := resp.ExtractInto(&latest); err != nil {
			return false, err
		}

		if latest.Status == status {
			// success case
			ret.Body = resp.Body
			return true, nil
		}

		if latest.Status == "ERROR" {
			errResp := latest.Error
			if errResp != nil {
				if errResp["details"] != nil {
					return false, fmt.Errorf(errResp["details"].(string))
				} else if errResp["message"] != nil {
					return false, fmt.Errorf(errResp["message"].(string))
				} else {
					return false, fmt.Errorf("Unknown error has occurred.")
				}
			}
		}

		return false, nil
	})
}
