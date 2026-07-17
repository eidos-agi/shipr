# Security

Report vulnerabilities to daniel@eidosagi.com.

Shipr writes local release memory under `.shipr/`. Do not store secrets,
credentials, raw tokens, private customer data, or unreleased sensitive details
in release models or release attempts. Store pointers to secret locations, not
secret values.

Shipr must stop before public release tags, package publishing, production
deployments, credential changes, payments, filings, and outbound announcements
unless the user explicitly approves the action.
