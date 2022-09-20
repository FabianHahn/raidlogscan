package pubsub

import (
	google_pubsub "cloud.google.com/go/pubsub"
)

type MessagePublishedData struct {
	Message google_pubsub.Message
}
