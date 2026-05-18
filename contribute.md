# Contributing

## Commit Message Convention

Use Linux/Git-style commit subjects with an area prefix:

```text
<area>: <summary>
```

The `<area>` should describe the most relevant affected scope, subsystem, package, directory, or component, such as `auth`, `api`, `ui`, `db`, `ci`, `docs`, `packages/core`, `packages/web`, or the changed module name.

Do not use Conventional Commits formats such as `feat:`, `fix:`, `chore:`, or `fix(auth)`.

Write the summary in imperative mood, keep it concise and specific, and keep the first line under 72 characters when possible. Use lowercase after the colon unless referring to proper nouns, symbols, APIs, or identifiers. Do not end the summary with a period.

Examples:

```text
auth: reject expired sessions
api: validate pagination params
ui: fix modal overflow
docs: clarify OpenAPI generation
web/api: add typed OpenAPI client
```

Prefer one-line commit messages for simple changes. For larger or multi-area changes, write a short subject line followed by a blank line and a body explaining what changed and why. Do not mention implementation details that are obvious from the diff.

## GitHub Copilot Commit Messages

If you use VS Code with GitHub Copilot, you can guide generated commit messages with this setting:

```json
{
	"github.copilot.chat.commitMessageGeneration.instructions": [
		{
			"text": "Generate Git commit messages using the Linux/Git-style area prefix format: <area>: <summary>. The <area> should describe the most relevant affected scope, subsystem, package, directory, or component, such as auth, api, ui, db, ci, docs, packages/core, packages/web, or the changed module name. Do not use Conventional Commits format such as feat:, fix:, chore:, or fix(auth). The summary must be concise, specific, and written in imperative mood, e.g. 'auth: reject expired sessions', 'api: validate pagination params', 'ui: fix modal overflow'. Keep the first line under 72 characters when possible. Use lowercase after the colon unless referring to proper nouns, symbols, APIs, or identifiers. Do not end the summary with a period. Prefer one-line commit messages for simple changes. For larger or multi-area changes, write a short subject line followed by a blank line and a body explaining what changed and why. Do not mention implementation details that are obvious from the diff."
		}
	]
}
```
