package console

import "strings"

// splitQueues turns a comma-separated option into a list of queue names.
func splitQueues(value string) []string {
	var queues []string

	for _, queue := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(queue); trimmed != "" {
			queues = append(queues, trimmed)
		}
	}

	return queues
}
