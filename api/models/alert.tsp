import "@typespec/http";

using Http;

namespace Netstamp;

alias AlertSeverity = "info" | "warning" | "critical";
alias AlertRuleStatus = "enabled" | "disabled";
alias AlertIncidentStatus = "open" | "acknowledged" | "resolved";
alias AlertEvaluationState = "firing" | "clear" | "insufficient_samples" | "no_data";
alias AlertConditionType = "metric_threshold";
alias AlertConditionMetric =
	| "ping.loss_percent"
	| "ping.average_rtt_ms"
	| "ping.max_rtt_ms"
	| "ping.success_rate"
	| "tcp.failure_percent"
	| "tcp.average_connect_ms"
	| "tcp.max_connect_ms"
	| "tcp.success_rate";
alias AlertConditionOperator = "gt" | "gte" | "lt" | "lte" | "eq";
alias NotificationChannelType = "webhook" | "discord" | "telegram";

model AlertRuleIdPathParam {
	@path
	rule_id: uuid;
}

model AlertIncidentIdPathParam {
	@path
	incident_id: uuid;
}

model NotificationChannelIdPathParam {
	@path
	channel_id: uuid;
}

@example(#{ status: "enabled", checkType: "ping" })
model AlertRuleListQuery {
	@query
	status?: AlertRuleStatus;

	@query
	checkType?: CheckType;
}

@example(#{ status: "open", limit: 50 })
model AlertIncidentListQuery {
	@query
	status?: AlertIncidentStatus;

	@query
	@minValue(1)
	@maxValue(200)
	limit?: int32 = 100;
}

@example(#{ checkType: "ping", probeId: "33333333-3333-3333-3333-333333333333", checkId: "44444444-4444-4444-4444-444444444444" })
model AlertRuleScope {
	checkType: CheckType;
	probeId?: uuid;
	checkId?: uuid;
}

@example(#{ type: "metric_threshold", metric: "ping.loss_percent", operator: "gte", threshold: 10, windowSeconds: 300, minSamples: 3 })
model AlertCondition {
	type: AlertConditionType;
	metric: AlertConditionMetric;
	operator: AlertConditionOperator;
	threshold: float64;

	@minValue(60)
	@maxValue(86400)
	windowSeconds: int32;

	@minValue(1)
	@maxValue(10000)
	minSamples: int32;
}

@example(#{
	id: "77777777-7777-7777-7777-777777777777",
	name: "API packet loss",
	description: "Page the team when API ping loss is high.",
	enabled: true,
	severity: "critical",
	scope: #{ checkType: "ping", probeId: "33333333-3333-3333-3333-333333333333", checkId: "44444444-4444-4444-4444-444444444444" },
	condition: #{ type: "metric_threshold", metric: "ping.loss_percent", operator: "gte", threshold: 10, windowSeconds: 300, minSamples: 3 },
	cooldownSeconds: 900,
	notificationChannelIds: #["88888888-8888-8888-8888-888888888888"],
	createdAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
	updatedAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
})
model AlertRule {
	id: uuid;

	@minLength(1)
	@maxLength(128)
	name: string;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	enabled: boolean;
	severity: AlertSeverity;
	scope: AlertRuleScope;
	condition: AlertCondition;

	@minValue(60)
	@maxValue(86400)
	cooldownSeconds: int32;

	notificationChannelIds: uuid[];
	createdAt: utcDateTime;
	updatedAt: utcDateTime;
}

@example(#{
	rule: #{
		id: "77777777-7777-7777-7777-777777777777",
		name: "API packet loss",
		enabled: true,
		severity: "critical",
		scope: #{ checkType: "ping" },
		condition: #{ type: "metric_threshold", metric: "ping.loss_percent", operator: "gte", threshold: 10, windowSeconds: 300, minSamples: 3 },
		cooldownSeconds: 900,
		notificationChannelIds: #[],
		createdAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
		updatedAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
	},
})
model AlertRuleResponse {
	rule: AlertRule;
}

@example(#{ rules: #[] })
model AlertRuleListResponse {
	rules: AlertRule[];
}

@example(#{
	name: "API packet loss",
	description: "Page the team when API ping loss is high.",
	enabled: true,
	severity: "critical",
	scope: #{ checkType: "ping" },
	condition: #{ type: "metric_threshold", metric: "ping.loss_percent", operator: "gte", threshold: 10, windowSeconds: 300, minSamples: 3 },
	cooldownSeconds: 900,
	notificationChannelIds: #["88888888-8888-8888-8888-888888888888"],
})
model CreateAlertRuleRequest {
	@minLength(1)
	@maxLength(128)
	name: string;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	enabled: boolean;
	severity: AlertSeverity;
	scope: AlertRuleScope;
	condition: AlertCondition;

	@minValue(60)
	@maxValue(86400)
	cooldownSeconds?: int32 = 900;

	notificationChannelIds?: uuid[];
}

@doc("Full replacement patch for an alert rule in the current beta API.")
model UpdateAlertRuleRequest is CreateAlertRuleRequest;

@example(#{
	state: "firing",
	metric: "ping.loss_percent",
	operator: "gte",
	threshold: 10,
	value: 12.5,
	samples: 5,
	minSamples: 3,
	windowSeconds: 300,
	windowStart: utcDateTime.fromISO("2026-06-14T08:55:00Z"),
	windowEnd: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
})
model AlertEvaluationSummary {
	state: AlertEvaluationState;
	metric: AlertConditionMetric;
	operator?: AlertConditionOperator;
	threshold?: float64;
	value?: float64;
	samples?: int64;
	minSamples?: int32;
	windowSeconds?: int32;
	windowStart?: utcDateTime;
	windowEnd?: utcDateTime;
}

@example(#{
	id: "99999999-9999-9999-9999-999999999999",
	ruleId: "77777777-7777-7777-7777-777777777777",
	probeId: "33333333-3333-3333-3333-333333333333",
	checkId: "44444444-4444-4444-4444-444444444444",
	checkType: "ping",
	status: "open",
	severity: "critical",
	lastEvaluationState: "firing",
	openedAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
	lastEvaluatedAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
	lastTriggeredAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
	lastValue: 12.5,
	lastSummary: #{ state: "firing", metric: "ping.loss_percent", value: 12.5, samples: 5 },
	suppressedNotificationCount: 0,
	createdAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
	updatedAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
})
model AlertIncident {
	id: uuid;
	ruleId: uuid;
	probeId: uuid;
	checkId: uuid;
	checkType: CheckType;
	status: AlertIncidentStatus;
	severity: AlertSeverity;
	lastEvaluationState: AlertEvaluationState;
	openedAt: utcDateTime;
	resolvedAt?: utcDateTime;
	lastEvaluatedAt: utcDateTime;
	lastTriggeredAt: utcDateTime;
	lastValue?: float64;
	lastSummary: AlertEvaluationSummary;
	lastNotificationSentAt?: utcDateTime;
	nextNotificationEligibleAt?: utcDateTime;
	suppressedNotificationCount: int32;
	createdAt: utcDateTime;
	updatedAt: utcDateTime;
}

@example(#{
	incident: #{
		id: "99999999-9999-9999-9999-999999999999",
		ruleId: "77777777-7777-7777-7777-777777777777",
		probeId: "33333333-3333-3333-3333-333333333333",
		checkId: "44444444-4444-4444-4444-444444444444",
		checkType: "ping",
		status: "open",
		severity: "critical",
		lastEvaluationState: "firing",
		openedAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
		lastEvaluatedAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
		lastTriggeredAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
		lastSummary: #{ state: "firing", metric: "ping.loss_percent" },
		suppressedNotificationCount: 0,
		createdAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
		updatedAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
	},
})
model AlertIncidentResponse {
	incident: AlertIncident;
}

@example(#{ incidents: #[] })
model AlertIncidentListResponse {
	incidents: AlertIncident[];
}

@example(#{ url: "https://hooks.example.com/netstamp" })
model WebhookChannelConfig {
	@format("uri")
	url: string;
}

@example(#{ url: "https://discord.com/api/webhooks/123/token" })
model DiscordChannelConfig {
	@format("uri")
	url: string;
}

@example(#{ botToken: "123456:telegram-bot-token", chatId: "-1001234567890" })
model TelegramChannelConfig {
	@minLength(1)
	botToken: string;

	@minLength(1)
	chatId: string;
}

@example(#{ chatId: "-1001234567890", botTokenConfigured: true })
model TelegramChannelResponseConfig {
	@minLength(1)
	chatId: string;

	botTokenConfigured: boolean;
}

alias NotificationChannelConfig = WebhookChannelConfig | DiscordChannelConfig | TelegramChannelResponseConfig;
alias CreateNotificationChannelConfig = WebhookChannelConfig | DiscordChannelConfig | TelegramChannelConfig;

@example(#{
	id: "88888888-8888-8888-8888-888888888888",
	name: "Incident webhook",
	type: "webhook",
	enabled: true,
	config: #{ url: "https://hooks.example.com/..." },
	createdAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
	updatedAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
})
model NotificationChannel {
	id: uuid;

	@minLength(1)
	@maxLength(128)
	name: string;

	type: NotificationChannelType;
	enabled: boolean;
	config: NotificationChannelConfig;
	createdAt: utcDateTime;
	updatedAt: utcDateTime;
}

@example(#{
	channel: #{
		id: "88888888-8888-8888-8888-888888888888",
		name: "Incident webhook",
		type: "webhook",
		enabled: true,
		config: #{ url: "https://hooks.example.com/..." },
		createdAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
		updatedAt: utcDateTime.fromISO("2026-06-14T09:00:00Z"),
	},
})
model NotificationChannelResponse {
	channel: NotificationChannel;
}

@example(#{ channels: #[] })
model NotificationChannelListResponse {
	channels: NotificationChannel[];
}

@example(#{ name: "Incident webhook", type: "webhook", enabled: true, config: #{ url: "https://hooks.example.com/netstamp" } })
model CreateNotificationChannelRequest {
	@minLength(1)
	@maxLength(128)
	name: string;

	type: NotificationChannelType;
	enabled: boolean;
	config: CreateNotificationChannelConfig;
}

@doc("Full replacement patch for a notification channel in the current beta API.")
model UpdateNotificationChannelRequest is CreateNotificationChannelRequest;

@example(#{ delivered: true, retryable: false, message: "Test notification delivered." })
model NotificationChannelTestResult {
	delivered: boolean;
	retryable: boolean;
	kind?: string;
	code?: string;
	message?: string;
}

@example(#{ result: #{ delivered: true, retryable: false, message: "Test notification delivered." } })
model NotificationChannelTestResponse {
	result: NotificationChannelTestResult;
}
