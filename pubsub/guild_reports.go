package pubsub

import (
	"fmt"
	"strconv"

	"github.com/cloudevents/sdk-go/v2/event"
)

func ParseGuildReportsEvent(e event.Event) (int64, error) {
	var message MessagePublishedData
	if err := e.DataAs(&message); err != nil {
		return 0, fmt.Errorf("failed to parse event message data: %v", err)
	}
	return strconv.ParseInt(message.Message.Attributes["guild_id"], 10, 64)
}
