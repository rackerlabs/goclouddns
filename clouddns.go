package goclouddns

import (
	"github.com/gophercloud/gophercloud/v2"
)

func NewCloudDNS(client *gophercloud.ProviderClient, eo gophercloud.EndpointOpts) (*gophercloud.ServiceClient, error) {

	sc := new(gophercloud.ServiceClient)

	serviceType := "rax:dns"
	eo.ApplyDefaults(serviceType)

	endpoint, err := client.EndpointLocator(eo)
	if err != nil {
		return sc, err
	}

	sc.ProviderClient = client
	sc.Endpoint = endpoint
	sc.Type = serviceType
	return sc, nil
}
