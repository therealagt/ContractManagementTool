# Verfahrensdokumentation — Contract Management Tool

Dokumentation für den Betrieb des NDA/AVV-Archivsystems auf GCP (Region `europe-west3`).

## Zweck

Revisionssichere Archivierung signierter NDAs und AVVs mit Human-in-the-Loop (HITL), hash-verkettetem Audit-Trail und WORM-Speicherung nach menschlicher Bestätigung.

## Rollen und Zugriff

| Rolle | Google Group (Beispiel) | Berechtigung |
|-------|----------------------|--------------|
| Uploader | `grp-contract-uploader@` | PDF-Upload |
| Reviewer | `grp-contract-reviewer@` | Review-Queue, Bestätigung/Ablehnung |
| Auditor | `grp-contract-auditor@` | Audit-Berichte, Legal Holds |
| Admin | `grp-contract-admin@` | Alle Rollen, Eskalation P1 |
| Ops | `grp-contract-ops@` | Operative Alerts (P1/P2) |

Zugriff ausschließlich über IAP + HTTPS Load Balancer. Keine öffentlichen Cloud-Run-URLs, keine user-facing GCS Signed URLs.

## Workflow (Kurz)

1. Upload → PAdES-Validierung (fail closed)
2. Staging → Gemini-Extraktion als **Draft**
3. HITL-Review → SoD: `uploaded_by ≠ confirmed_by`
4. Archivierung in WORM-Bucket nach Bestätigung
5. Nächtlicher Integritäts-Cron (02:00 Europe/Berlin)
6. Wöchentlicher Statusbericht (Montag 08:00 Europe/Berlin)

## Access Review (vierteljährlich)

1. **IAP-Gruppen:** Export der Mitglieder aller `grp-contract-*` Gruppen; Abgleich mit HR/IT-Liste aktiver Mitarbeiter.
2. **API-Rollen:** Prüfung der `AUTH_*_EMAILS` Terraform-Variablen gegen tatsächliche Zuweisungen.
3. **Service Accounts:** Keine User-Keys; nur Workload-Identities und Scheduler-OIDC.
4. **Dokumentation:** Ergebnis im Ticket-System (z. B. Jira) mit Datum, Prüfer, Abweichungen und Maßnahmen archivieren.
5. **Entzug:** Entfernte Nutzer aus IAP-Gruppen und `auth_*_emails` innerhalb von 24 h nach Offboarding.

## Alerting und Eskalation

| Severity | Trigger | Empfänger | Reaktionszeit |
|----------|---------|-----------|---------------|
| P1 | Hash-Abweichung, Audit-Chain gebrochen | Ops + nach 4 h Admin | Sofort |
| P2 | Review-SLA, DLQ, Extraktionsfehler | Ops | Sofort |
| P3 | Wöchentlicher Statusbericht | Ops + Audit | Wöchentlich |

Alle Alerts werden in `alert_events` (PostgreSQL + BigQuery) persistiert.

---

## P1 Incident Runbook

**Auslöser:** Integritäts-Cron meldet SHA-256-Abweichung oder Audit-Hash-Chain ungültig.

### Sofort (0–15 Min)

1. **Incident eröffnen** — Ticket mit Severity P1, Zeitstempel, betroffene `contract_id`(s).
2. **Kein Auto-Repair** — WORM-Objekte nicht überschreiben, löschen oder „reparieren“.
3. **Cloud Monitoring** — Incident und betroffene Metrik (`integrity_check_failed` / `audit_chain_broken`) prüfen.
4. **IAP-Dashboard** — Audit-Trail und Contract-Status für betroffene IDs einsehen.

### Analyse (15–60 Min)

1. **Erwarteter vs. aktueller Hash** aus Alert-Payload / `integrity_check_runs` / `archive_records` notieren.
2. **GCS-Objekt** — Pfad aus `archive_records.gcs_path`; Objekt-Metadaten und Versionen prüfen (nur lesend).
3. **Staging** — Prüfen ob Staging-Objekt noch existiert (sollte nach Archivierung gelöscht sein).
4. **Audit-Events** — Letzte Aktionen (`archived`, `upload_rejected`, `legal_hold`) für die Vertrags-ID.

### Eskalation (nach 4 h ohne Lösung)

- Automatische E-Mail an `grp-contract-admin@` via Cloud Monitoring Escalation Policy.
- Admin informiert Compliance / Legal bei Verdacht auf Datenmanipulation oder Archivfehler.

### Abschluss

1. **Root Cause** dokumentieren (technisch, prozessual, menschlich).
2. **Corrective Action** — z. B. Wiederherstellung aus DocuSign-Original, manuelle Neuarchivierung nach neuem HITL-Zyklus (nie stilles Überschreiben).
3. **Incident schließen** in Cloud Monitoring → stoppt Eskalation.
4. **Audit-Eintrag** — `alert_events` und ggf. manuelles `audit_events` mit `action=incident_resolved`.

### Verboten

- Automatisches Reparieren von Hash-Mismatches
- Löschen von WORM-Archivobjekten ohne Legal/Compliance-Freigabe
- Umgehen von IAP oder SoD zur „schnellen“ Behebung

---

## Wöchentlicher Statusbericht

Cloud Scheduler → `contract-weekly-report` (internal only) → HTML-E-Mail an Ops + Audit.

Inhalt: Ampel-Status, Pipeline-Kennzahlen der Woche, Integritätsstatus, P1/P2-Alerts, offene Compliance-Lücken. Kein Vertragstext, keine PII.

Formeller Audit-Bericht (PDF/Evidence Package) bleibt on-demand über IAP (`/audit/*`).

## Änderungsmanagement

- Infrastruktur: Terraform (`terraform/environments/`), PR + Review
- Anwendung: GitHub Actions (WIF), Binary Authorization in Prod (nur Images aus Projekt-AR)
- Prod Bucket Lock: irreversibel — erst nach Validierung in Dev/Staging aktivieren

## Kontakt

| Bereich | Gruppe |
|---------|--------|
| Operativ | `grp-contract-ops@` |
| Eskalation P1 | `grp-contract-admin@` |
| Audit | `grp-contract-auditor@` |
