import "@typespec/http";

using Http;

namespace Netstamp;

@format("uuid")
scalar uuid extends string;

@format("email")
scalar email extends string;

@format("ipv4")
scalar ipv4Address extends string;

@format("ipv6")
scalar ipv6Address extends string;

alias ipAddress = ipv4Address | ipv6Address;

alias ProjectRef = string;

alias SelectorLabelOperator = "eq" | "in" | "exists";

@example(#{ key: "region", `op`: "eq", value: "tokyo" })
model SelectorLabelExpression {
	@minLength(1)
	@maxLength(128)
	key: string;

	`op`: SelectorLabelOperator;

	@minLength(1)
	@maxLength(256)
	value?: string;

	@minItems(1)
	values?: string[];
}

@example(#{ label: #{ key: "region", `op`: "eq", value: "tokyo" } })
@doc("Selector AST used to match probes by project labels. Empty object matches all probes.")
model Selector {
	all?: Selector[];
	any?: Selector[];
	not?: Selector;
	label?: SelectorLabelExpression;
}

model SessionCookieAuth {
	type: AuthType.apiKey;
	in: ApiKeyLocation.cookie;
	name: "__Host-netstamp_session";
}

@doc("Probe runtime credential header. Send the full value as Authorization: Probe <secret>.")
model ProbeAuth {
	type: AuthType.apiKey;
	in: ApiKeyLocation.header;
	name: "Authorization";
}

alias ErrorCode =
	"BAD_REQUEST" |
	"UNAUTHORIZED" |
	"FORBIDDEN" |
	"NOT_FOUND" |
	"CONFLICT" |
	"RATE_LIMITED" |
	"VALIDATION_FAILED" |
	"INTERNAL_ERROR" |
	"SERVICE_UNAVAILABLE" |
	"METHOD_NOT_ALLOWED" |
	"ROUTE_NOT_FOUND" |
	"READ_ONLY" |
	"AUTH_MISSING_SESSION" |
	"AUTH_INVALID_SESSION" |
	"AUTH_INVALID_CSRF" |
	"AUTH_INVALID_CREDENTIALS" |
	"AUTH_EMAIL_VERIFICATION_REQUIRED" |
	"AUTH_REGISTRATION_DISABLED" |
	"AUTH_EMAIL_VERIFICATION_UNAVAILABLE" |
	"AUTH_PASSWORD_RESET_UNAVAILABLE" |
	"AUTH_EMAIL_VERIFICATION_TOKEN_INVALID" |
	"AUTH_PASSWORD_RESET_TOKEN_INVALID" |
	"AUTH_SESSION_NOT_FOUND" |
	"USER_NOT_FOUND" |
	"EMAIL_ALREADY_EXISTS" |
	"LAST_SYSTEM_ADMIN" |
	"SYSTEM_ADMIN_REQUIRED" |
	"SELF_SYSTEM_ADMIN_ACTION" |
	"INVALID_ADMIN_DATA_IMPORT" |
	"PROJECT_ROLE_REQUIRED" |
	"PROJECT_NOT_FOUND" |
	"PROJECT_SLUG_ALREADY_EXISTS" |
	"PROJECT_MEMBER_NOT_FOUND" |
	"PROJECT_MEMBER_ALREADY_EXISTS" |
	"PROJECT_INVITE_NOT_FOUND" |
	"PROJECT_INVITE_ALREADY_EXISTS" |
	"PROJECT_LAST_OWNER" |
	"LABEL_NOT_FOUND" |
	"LABEL_ALREADY_EXISTS" |
	"CHECK_NOT_FOUND" |
	"PROBE_NOT_FOUND" |
	"PROBE_CREDENTIAL_INVALID" |
	"PROBE_DISABLED" |
	"PROBE_RUNTIME_AUTH_UNAVAILABLE" |
	"ALERT_RULE_NOT_FOUND" |
	"ALERT_INCIDENT_NOT_FOUND" |
	"ALERT_NOTIFICATION_NOT_FOUND" |
	"PUBLIC_STATUS_PAGE_NOT_FOUND" |
	"PUBLIC_STATUS_ELEMENT_NOT_FOUND" |
	"PUBLIC_STATUS_SLUG_ALREADY_EXISTS" |
	"AGENT_BINARY_NOT_FOUND" |
	"AGENT_BINARY_UNAVAILABLE";

alias ProblemFieldErrorCode =
	"INVALID_VALUE" |
	"INVALID_FORMAT" |
	"INVALID_ENUM" |
	"REQUIRED" |
	"VALUE_TOO_SHORT" |
	"VALUE_TOO_LONG" |
	"VALUE_TOO_SMALL" |
	"VALUE_TOO_LARGE";

@example(#{ code: "REQUIRED", message: "expected required property name to be present", location: "body.name", value: "" })
model ProblemErrorDetail {
	@doc("Stable machine-readable code for this field or sub-error.")
	code?: ProblemFieldErrorCode;

	@doc("Human-readable explanation for this field or sub-error.")
	message?: string;

	@doc("Path-like location such as body.name, query.probeId, or path.ref.")
	location?: string;

	@doc("Echoed invalid value, when the server chooses to return it.")
	value?: unknown;
}

@error
@example(#{ code: "VALIDATION_FAILED", title: "Unprocessable Entity", status: 422, detail: "invalid project input", requestId: "req_01J00000000000000000000000", errors: #[#{ code: "VALUE_TOO_SHORT", message: "must be at least 1 character", location: "body.name", value: "" }] })
model ProblemDetails {
	@format("uri")
	type?: string = "about:blank";

	@doc("Stable machine-readable problem code. Clients should branch on this field instead of title or detail.")
	code?: ErrorCode;

	title?: string;
	status?: int32;
	detail?: string;

	@format("uri")
	instance?: string;

	requestId?: string;

	errors?: ProblemErrorDetail[];
}

model OkJson<T> {
	...OkResponse;
	...Body<T>;
}

model CreatedJson<T> {
	...CreatedResponse;
	...Body<T>;
}

model Accepted {
	...Response<202>;
}

model AcceptedJson<T> {
	...Response<202>;
	...Body<T>;
}

model NoContent {
	...NoContentResponse;
}

model SessionOkJson<T> {
	...OkResponse;

	@header("Set-Cookie")
	setCookie: string;

	...Body<T>;
}

model SessionCreatedJson<T> {
	...CreatedResponse;

	@header("Set-Cookie")
	setCookie: string;

	...Body<T>;
}

model SessionNoContent {
	...NoContentResponse;

	@header("Set-Cookie")
	setCookie: string;
}

model OptionalSessionNoContent {
	...NoContentResponse;

	@header("Set-Cookie")
	setCookie?: string;
}

model BadRequestProblem {
	...BadRequestResponse;
	...ProblemJson;
}

model UnauthorizedProblem {
	...UnauthorizedResponse;
	...ProblemJson;
}

model ForbiddenProblem {
	...ForbiddenResponse;
	...ProblemJson;
}

model NotFoundProblem {
	...NotFoundResponse;
	...ProblemJson;
}

model ConflictProblem {
	...ConflictResponse;
	...ProblemJson;
}

model TooManyRequestsProblem {
	...Response<429>;
	...ProblemJson;
}

model UnprocessableEntityProblem {
	...Response<422>;
	...ProblemJson;
}

model InternalServerProblem {
	...Response<500>;
	...ProblemJson;
}

model ServiceUnavailableProblem {
	...Response<503>;
	...ProblemJson;
}

model ProblemJson {
	@header
	contentType: "application/problem+json";

	...Body<ProblemDetails>;
}

model ProjectRefParam {
	@path
	@minLength(1)
	@maxLength(64)
	@pattern("^[a-z0-9-]+$")
	ref: ProjectRef;
}

model ProbeIdPathParam {
	@path
	probe_id: uuid;
}

model CheckIdPathParam {
	@path
	check_id: uuid;
}

model LabelIdPathParam {
	@path
	label_id: uuid;
}

model UserIdPathParam {
	@path
	user_id: uuid;
}

model ProjectInviteIdPathParam {
	@path
	invite_id: uuid;
}
