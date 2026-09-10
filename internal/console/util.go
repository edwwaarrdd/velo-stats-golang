package console

import "strings"

func splitQueues(value string) []string {
	var queues []string

	for _, queue := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(queue); trimmed != "" {
			queues = append(queues, trimmed)
		}
	}

	return queues
}
