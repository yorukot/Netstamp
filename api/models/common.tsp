import "@typespec/http";

using Http;

namespace Netstamp;

@format("uuid")
scalar uuid extends string;

@format("email")
scalar email extends string;

scalar ipAddress extends string;

alias ProjectRef = string;

model SessionCookieAuth {
	type: AuthType.apiKey;
	in: ApiKeyLocation.cookie;
	name: "netstamp_session";
}

model ProbeAuth {
	type: AuthType.http;
	scheme: "Probe";
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

model UnauthorizedProblem {
	...UnauthorizedResponse;
	...Body<ProblemDetails>;
}

model ForbiddenProblem {
	...ForbiddenResponse;
	...Body<ProblemDetails>;
}

model NotFoundProblem {
	...NotFoundResponse;
	...Body<ProblemDetails>;
}

model ConflictProblem {
	...ConflictResponse;
	...Body<ProblemDetails>;
}

model UnprocessableEntityProblem {
	...Response<422>;
	...Body<ProblemDetails>;
}

model InternalServerProblem {
	...Response<500>;
	...Body<ProblemDetails>;
}

model ServiceUnavailableProblem {
	...Response<503>;
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
