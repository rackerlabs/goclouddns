# goclouddns

A [gophercloud][gophercloud] compatible Go module for supporting
[Rackspace Cloud DNS][raxclouddns]. If you use [goraxauth][goraxauth],
like the test binary does then you can support Rackspace API keys.

## Install

```bash
go install github.com/rackerlabs/goclouddns/cmd/clouddns@latest
```

You can also download prebuilt binaries from
[GitHub Releases](https://github.com/rackerlabs/goclouddns/releases).

## CLI Usage

To use the CLI, you'll need a cloud username and password or API key.

```bash
export OS_USERNAME=[username]
export RAX_API_KEY=[api-key]
# or if you can auth by a password
export OS_PASSWORD=[password]
clouddns domain list
clouddns record list <domain-id>
```

[gophercloud]: <https://github.com/gophercloud/gophercloud>
[goraxauth]: <https://github.com/rackerlabs/goraxauth>
[raxclouddns]: <https://docs.rackspace.com/docs/cloud-dns>
