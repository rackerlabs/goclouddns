# goclouddns

A [gophercloud][gophercloud] compatible Go module for supporting
[Rackspace Cloud DNS][raxclouddns]. If you use [goraxauth][goraxauth],
like the test binary does then you can support Rackspace API keys.

To run the test binary, you'll need a cloud username and password/api-key.

```bash
export OS_USERNAME=[username]
export OS_TENANT_ID=[tenantid]
# password or API key
export OS_PASSWORD=[password]
export RAX_API_KEY=[api-key]
go run ./cmd/clouddns domain list
```

[gophercloud]: <https://github.com/gophercloud/gophercloud>
[goraxauth]: <https://github.com/rackerlabs/goraxauth>
[raxclouddns]: <https://docs.rackspace.com/docs/cloud-dns>
