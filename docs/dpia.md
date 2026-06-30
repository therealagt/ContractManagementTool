# Datenschutz-Folgenabschätzung (DPIA) — Contract Management Tool

**Stand:** Entwurf für interne Freigabe  
**Verantwortlich:** Datenschutzbeauftragter / Compliance  
**System:** NDA/AVV-Archiv auf GCP (`europe-west3`)

## 1. Beschreibung der Verarbeitung

Das System archiviert **signierte NDAs und AVVs** (PDF) und extrahiert strukturierte Metadaten (Vertragsparteien, Laufzeiten, Verarbeitungszwecke) zur Nachweisführung gegenüber Auditoren und Aufsicht.

| Phase | Verarbeitung | Speicherort |
|-------|--------------|-------------|
| Upload | PAdES-Validierung, Staging | GCS Staging (CMEK) |
| Extraktion | Gemini-Extraktion als Draft | PostgreSQL `extraction_draft` |
| Review | Menschliche Bestätigung (HITL) | PostgreSQL `confirmed_metadata` |
| Archiv | WORM nach Bestätigung | GCS Archive (Retention Lock prod) |
| Audit | Hash-Chain Events, Access Log | PostgreSQL + BigQuery |

**Grundsatz Datenminimierung:** Kein Volltext der Verträge in DB/BigQuery — nur GCS. E-Mail-Statusberichte enthalten aggregierte Kennzahlen und Vertrags-IDs, keine Vertragstexte.

## 2. Zweck und Rechtsgrundlage

| Zweck | Rechtsgrundlage (DSGVO) |
|-------|-------------------------|
| Nachweis von Geheimhaltungs- und Auftragsverarbeitungsverträgen | Art. 6 Abs. 1 lit. f (berechtigtes Interesse) / lit. b |
| Erfüllung von Art. 28 DSGVO (AVV-Dokumentation) | Art. 6 Abs. 1 lit. c |
| Audit- und Compliance-Nachweise (ISO 27001) | Art. 6 Abs. 1 lit. f |

## 3. Betroffene Personen und Datenkategorien

- **Mitarbeiter / Geschäftspartner** als Vertragsparteien in Metadaten (Name, Firma, ggf. E-Mail in PDF)
- **Interne Nutzer** (Uploader, Reviewer) in Audit-Events (`actor` = IAP-E-Mail)
- **Keine** besonderen Kategorien nach Art. 9 DSGVO im Regelbetrieb (sofern Verträge keine solchen Daten enthalten)

## 4. Empfänger und Drittlandtransfer

| Empfänger | Rolle | Region |
|-----------|-------|--------|
| Google Cloud (GCP) | Auftragsverarbeiter | `europe-west3` (Frankfurt) |
| Vertex AI (Gemini) | KI-Extraktion (Draft only) | EU (`europe-west3`), Enterprise Terms |

Kein Training auf Kundendaten. VPC Service Controls optional für Exfiltrationsschutz.

## 5. Technische und organisatorische Maßnahmen

- IAP + HTTPS LB, Cloud Armor (Prod IP-Allowlist)
- CMEK (KMS) für GCS, Cloud SQL, BigQuery
- WORM-Archiv, Legal Hold, Integritäts-Cron ohne Auto-Repair
- SoD: Upload ≠ Bestätigung
- Hash-verkettete `audit_events`, `access_events`
- Binary Authorization (Prod), keine SA-Keys
- Secrets nur in Secret Manager

## 6. Risiken und Maßnahmen

| Risiko | Eintritt | Auswirkung | Maßnahme |
|--------|----------|------------|----------|
| Unbefugter Zugriff auf Verträge | Niedrig | Hoch | IAP, Rollen, Access Review |
| KI-Fehler in Metadaten | Mittel | Mittel | HITL-Pflicht, Draft ≠ Wahrheit |
| Datenverlust / Manipulation Archiv | Niedrig | Hoch | WORM, Hash-Checks, P1-Runbook |
| Übermittlung an Drittländer | Niedrig | Hoch | EU-Region, Vertex EU, AVV mit Google |
| PII in Alerts/E-Mails | Niedrig | Mittel | Nur Aggregat + IDs in Weekly Report |

## 7. Notwendigkeit und Verhältnismäßigkeit

Die Verarbeitung ist erforderlich für nachweisbare Vertragsarchivierung und regulatorische Anforderungen (DSGVO Art. 28, ISO 27001). Alternativen (manueller SharePoint ohne WORM/Audit) wurden als unzureichend für Compliance-Anforderungen bewertet.

## 8. Betroffenenrechte

Auskunft/Löschung über Standard-Prozesse des Datenschutzteams. **Hinweis:** WORM-Archiv und Legal Hold können Löschung verzögern oder erfordern dokumentierte Ausnahme nach Legal Hold Release.

## 9. Review-Zyklus

- **Jährlich:** DPIA-Update bei wesentlichen Architekturänderungen
- **Vierteljährlich:** Access Review (siehe `verfahrensdokumentation.md`)
- **Bei Incident P1:** DPIA-Risikobewertung prüfen

## 10. Freigabe

| Rolle | Name | Datum | Unterschrift |
|-------|------|-------|--------------|
| Datenschutzbeauftragter | | | |
| IT-Security | | | |
| Fachverantwortlicher | | | |
