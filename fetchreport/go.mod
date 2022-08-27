module github.com/FabianHahn/raidlogscan/fetchreport

go 1.16

replace github.com/shurcooL/graphql => github.com/navops/graphql v0.0.2

require (
	cloud.google.com/go/datastore v1.8.0
	cloud.google.com/go/pubsub v1.24.0
	github.com/GoogleCloudPlatform/functions-framework-go v1.6.1
	github.com/cloudevents/sdk-go/v2 v2.6.1
	github.com/google/uuid v1.3.0 // indirect
	github.com/shurcooL/graphql v0.0.0-20220606043923-3cf50f8a0a29
	github.com/stretchr/testify v1.7.1 // indirect
	golang.org/x/oauth2 v0.0.0-20220808172628-8227340efae7
)
