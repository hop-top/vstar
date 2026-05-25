# V* — Glossary

| Term | Definition |
|---|---|
| **V\*** | Convention for representing agentic-system state as iCalendar (RFC 5545) + vCard (RFC 6350) components. Pronounced "vee-star". |
| **Component** | An iCalendar/vCard top-level entity: VCALENDAR, VTODO, VEVENT, VJOURNAL, VFREEBUSY, VTIMEZONE, VALARM, VCARD. |
| **Property** | A name/value pair inside a component (e.g. `UID:foo`, `DTSTAMP:20260504T180000Z`). |
| **Parameter** | A name/value pair on a property (e.g. `TZID=America/Montreal` in `DTSTART;TZID=America/Montreal:20260504T140000`). |
| **Canonical form** | The deterministic byte sequence for a component, used for equality + hashing. See `spec/03-canonicalization.md`. |
| **Supersession** | Append-only state change pattern using `RELATED-TO` + `X-VSTAR-EFFECTIVE-STATUS`. See `spec/02-component-mapping.md`. |
| **Emitter** | Implementation that produces V\* documents. |
| **Consumer** | Implementation that reads V\* documents and projects state. |
| **Round-trip** | Emitter + Consumer with byte-identical re-emission of self-emitted documents. |
| **Ledger** | An append-only sequence of V\* components. Storage format is consumer-defined (JSONL, SQLite, …); V\* defines only the components themselves. |
| **Projection** | The act of folding a ledger into current state. Consumer-defined; outside V\*'s scope. |
