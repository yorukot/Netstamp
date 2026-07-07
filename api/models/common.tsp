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
	name: "netstamp_session";
}

@doc("Probe runtime credential header. Send the full value as Authorization: Probe <secret>.")
model ProbeAuth {
	type: AuthType.apiKey;
	in: ApiKeyLocation.header;
	name: "Authorization";
}

@example(#{ message: "expected required property name to be present", location: "body.name", value: "" })
model ProblemErrorDetail {
	@doc("Human-readable explanation for this field or sub-error.")
	message?: string;

	@doc("Path-like location such as body.name, query.probeId, or path.ref.")
	location?: string;

	@doc("Echoed invalid value, when the server chooses to return it.")
	value?: unknown;
}

@error
@example(#{ title: "Unprocessable Entity", status: 422, detail: "invalid project input", errors: #[#{ message: "must be at least 1 character", location: "body.name", value: "" }] })
model ProblemDetails {
	@format("uri")
	type?: string = "about:blank";

	title?: string;
	status?: int32;
	detail?: string;

	@format("uri")
	instance?: string;

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
