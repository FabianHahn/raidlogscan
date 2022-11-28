package pubsub

import (
	"fmt"
	"strconv"

	"github.com/cloudevents/sdk-go/v2/event"
)

func ParseUserReportsEvent(e event.Event) (int32, error) {
	var message MessagePublishedData
	if err := e.DataAs(&message); err != nil {
		return 0, fmt.Errorf("failed to parse event message data: %v", err)
	}
	userId64, err := strconv.ParseInt(message.Message.Attributes["user_id"], 10, 64)
	if err != nil {
		return 0, err
	}
	return int32(userId64), nil
}
