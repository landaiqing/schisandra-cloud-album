package events

import (
	"fmt"
)

func LogHandler(event Event) {
	fmt.Printf("[LOG] Event: %s, Data: %+v\n", event.Name, event.Data)
}

func NotifyHandler(event Event) {
	fmt.Printf("[NOTIFY] User notified about event: %s\n", event.Name)
}
