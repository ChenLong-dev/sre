package resp

type GetK8sResourceEventsResp struct {
	Events []Event `json:"events" deepcopy:"from:Items"`
}

type Event struct {
	Reason         string      `json:"reason"`
	Message        string      `json:"message"`
	Source         EventSource `json:"source"`
	FirstTimestamp string      `json:"first_timestamp"`
	LastTimestamp  string      `json:"last_timestamp"`
	EventTime      string      `json:"event_time"`
	Type           string      `json:"type"`
	Count          int32       `json:"count"`
}

type EventSource struct {
	Component string `json:"component,omitempty"`
	Host      string `json:"host,omitempty"`
}
