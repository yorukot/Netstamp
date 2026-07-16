import "@typespec/http";

using Http;

namespace Netstamp;

alias APITokenScope =
	| "projects:read"
	| "projects:write"
	| "probes:read"
	| "probes:write"
	| "checks:read"
	| "checks:write"
	| "labels:read"
	| "labels:write"
	| "assignments:read"
	| "results:read"
	| "alerts:read"
	| "alerts:write"
	| "status_pages:read"
	| "status_pages:write";

model APIToken {
	id: uuid;
	@minLength(1) @maxLength(100) name: string;
	@minLength(8) @maxLength(8) tokenHint: string;
	@minItems(1) scopes: APITokenScope[];
	createdAt: utcDateTime;
	lastUsedAt?: utcDateTime;
	expiresAt: utcDateTime;
}

model CreateAPITokenRequest {
	@minLength(1) @maxLength(100) name: string;
	@minItems(1) scopes: APITokenScope[];
	expiresAt: utcDateTime;
}

model CreateAPITokenResponse {
	token: APIToken;

	@doc("One-time plaintext token. This value is never returned again.")
	value: string;
}

model APITokenListResponse {
	tokens: APIToken[];
}

model APITokenIdPathParam {
	@path("token_id") tokenId: uuid;
}
