To test this library, you'll need a cloud username and passwords - api keys do not work

```
export OS_AUTH_URL=https://identity.api.rackspacecloud.com/v2.0/
export OS_USERNAME=[username]
export OS_TENANT_ID=[tenantid]
export OS_PASSWORD=[password]
go run ./cmd/clouddns domain list
```
