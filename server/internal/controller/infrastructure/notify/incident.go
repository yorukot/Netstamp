package notify

import (
	"encoding/json"
	"fmt"
	"strings"
)

const incidentDescriptionLimit = 1024

type incidentNotificationView struct {
	EventType string `json:"eventType"`
	SentAt    string `json:"sentAt"`
	Rule      struct {
		Name     string `json:"name"`
		Severity string `json:"severity"`
	} `json:"rule"`
	Target struct {
		ProbeID   string `json:"probeId"`
		CheckID   string `json:"checkId"`
		CheckType string `json:"checkType"`
		Probe     struct {
			Name string `json:"name"`
		} `json:"probe"`
		Check struct {
			Name   string `json:"name"`
			Type   string `json:"type"`
			Target string `json:"target"`
		} `json:"check"`
	} `json:"target"`
	Notification struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"notification"`
	Links struct {
		Incident string `json:"incident"`
	} `json:"links"`
	Summary map[string]any `json:"summary"`
}

type incidentNotificationField struct {
	Name   string
	Value  string
	Inline bool
}

func parseIncidentNotificationPayload(payload []byte) (incidentNotificationView, bool) {
	var incident incidentNotificationView
	if err := json.Unmarshal(payload, &incident); err != nil {
		return incidentNotificationView{}, false
	}
	if incident.Summary == nil {
		incident.Summary = map[string]any{}
	}
	return incident, true
}

func incidentNotificationTitle(incident incidentNotificationView) string {
	if incident.EventType == "notification.test" {
		return "Netstamp test notification"
	}
	return "Netstamp alert"
}

func incidentNotificationDescription(incident incidentNotificationView) string {
	if message, ok := incident.Summary["message"].(string); ok && message != "" {
		return truncateMessage(message, incidentDescriptionLimit)
	}
	if incident.Target.CheckType != "" {
		return strings.ToUpper(incident.Target.CheckType) + " alert"
	}
	return ""
}

func incidentNotificationFields(incident incidentNotificationView) []incidentNotificationField {
	fields := []incidentNotificationField{}
	add := func(name, value string, inline bool) {
		if value == "" {
			return
		}
		fields = append(fields, incidentNotificationField{Name: name, Value: value, Inline: inline})
	}

	add("Rule", incident.Rule.Name, true)
	add("Severity", incident.Rule.Severity, true)
	add("Event", incident.EventType, true)
	add("Incident", incident.Links.Incident, false)
	add("Probe", incidentProbeLabel(incident), true)
	add("Check", incidentCheckLabel(incident), true)
	add("Target", incident.Target.Check.Target, true)
	if metric, ok := incident.Summary["metric"].(string); ok {
		add("Metric", metric, true)
	}
	if value, ok := incident.Summary["value"]; ok && value != nil {
		add("Value", fmt.Sprintf("%v", value), true)
	}
	if threshold, ok := incident.Summary["threshold"]; ok && threshold != nil {
		add("Threshold", fmt.Sprintf("%v", threshold), true)
	}
	add("Notification", incident.Notification.Name, true)
	add("Sent", incident.SentAt, false)
	return fields
}

func incidentProbeLabel(incident incidentNotificationView) string {
	if incident.Target.Probe.Name != "" {
		return incident.Target.Probe.Name
	}
	return incident.Target.ProbeID
}

func incidentCheckLabel(incident incidentNotificationView) string {
	name := incident.Target.Check.Name
	if name == "" {
		name = incident.Target.CheckID
	}
	checkType := incident.Target.Check.Type
	if checkType == "" {
		checkType = incident.Target.CheckType
	}
	if checkType == "" {
		return name
	}
	return name + " (" + strings.ToUpper(checkType) + ")"
}
