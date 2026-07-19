# Security policy

Netstamp handles network targets, probe credentials, user identities, notification secrets, and monitoring history. Please report suspected vulnerabilities privately so maintainers have a chance to investigate and coordinate a fix before technical details become public.

## Supported versions

The project currently develops and publishes from `main` and does not promise security backports for older images or commits. Reproduce against the latest available revision when it is safe to do so, and include the exact image tag, digest, or commit in your report.

## Report a vulnerability

Use [GitHub private vulnerability reporting](https://github.com/yorukot/netstamp/security/advisories/new) when it is available for the repository. Do not open a public issue or discussion containing exploit details, credentials, personal data, private monitoring payloads, or the location of an unpatched deployment.

Include as much of the following as possible:

- affected component, endpoint, and version or commit;
- prerequisites and the attacker's required access;
- reproducible steps or a minimal proof of concept;
- observed and expected security boundary;
- realistic impact and affected data;
- relevant logs with every secret and personal value removed;
- suggested mitigation, if known;
- whether the report or related details have been shared elsewhere.

If private reporting is unavailable, open a minimal public issue asking maintainers for a private contact channel. Do not include vulnerability details in that issue.

## What happens next

Maintainers will try to acknowledge a complete report, validate its impact, prepare a fix, and coordinate disclosure as availability permits. Netstamp is community maintained and does not guarantee a response or remediation SLA. Please avoid testing against systems you do not own or have explicit permission to assess.

After a fix is available, the project may publish a GitHub Security Advisory with affected versions, impact, mitigations, and credit. Tell the maintainers if you prefer not to be credited.

## Operational security

Deployment hardening, secret rotation, probe trust, public status privacy, and incident-response guidance live in the [Netstamp security documentation](https://netstamp.dev/docs/operate/security/).
