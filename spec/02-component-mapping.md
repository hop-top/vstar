# V* — Component Mapping

## Core mapping

| Agentic concept | V\* component | RFC |
|---|---|---|
| World | VCALENDAR | RFC 5545 |
| Environment | VJOURNAL or VCALENDAR category | RFC 5545 |
| Player | VCARD | RFC 6350 |
| Group | VCARD `KIND:group` | RFC 6350 |
| Mission | VTODO | RFC 5545 |
| Assignment | VTODO | RFC 5545 |
| Turn | VEVENT | RFC 5545 |
| Playthrough | VEVENT | RFC 5545 |
| Action | VJOURNAL | RFC 5545 |
| Observation | VJOURNAL | RFC 5545 |
| Decision | VJOURNAL | RFC 5545 |
| Score | VJOURNAL | RFC 5545 |
| Memory | VJOURNAL | RFC 5545 |
| Artifact | VJOURNAL + ATTACH | RFC 5545 |
| Resource delta | VJOURNAL | RFC 5545 |
| Learning | VJOURNAL | RFC 5545 |
| Availability | VFREEBUSY | RFC 5545 |
| Timeout / escalation | VALARM | RFC 5545 |
| Timezone | VTIMEZONE | RFC 5545 |

## Required common properties

Every V\* component MUST carry:

```text
UID
DTSTAMP
X-VSTAR-HASH
```

`X-VSTAR-HASH` is the canonical content hash defined in
`03-canonicalization.md`.

## Recommended properties

These are not required but strongly encouraged:

```text
CREATED
LAST-MODIFIED
SEQUENCE
RELATED-TO
CATEGORIES
```

`RELATED-TO` is the primary mechanism for expressing
component-to-component relationships (assignment → mission, action
→ turn, observation → action, etc.).

## Status supersession (append-only ledger)

V\* is append-only. Original components MUST NOT be mutated. State
changes are expressed as superseding VJOURNAL entries that
reference the original via `RELATED-TO`:

```text
BEGIN:VJOURNAL
UID:journal:status:assignment:todo-factory:scaffold-app:2026-05-04T18:00:00Z
DTSTAMP:20260504T180000Z
RELATED-TO:assignment:todo-factory:scaffold-app
CATEGORIES:status-supersession
X-VSTAR-EFFECTIVE-STATUS:COMPLETED
X-VSTAR-HASH:<canonical-hash>
END:VJOURNAL
```

Projection (state-from-log) is performed by consumers; V\* itself
defines only the supersession encoding.
