module github.com/rackerlabs/goclouddns

go 1.21.6

require (
	github.com/gophercloud/gophercloud v1.11.0
	github.com/rackerlabs/goraxauth v0.0.0-20240414034322-f28547de0fe7
)

replace github.com/gophercloud/gophercloud => github.com/cardoe/gophercloud v1.5.1-0.20240414024906-92421bde3201
