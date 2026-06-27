import hashlib
import json
from datetime import datetime, timezone
from typing import Any
from uuid import uuid4

from sqlalchemy import select
from sqlalchemy.orm import Session

from libs.common.db.models import AccessEvent, AuditEvent


class AuditChainError(ValueError):
    pass


def _hash_payload(payload: dict[str, Any]) -> str:
    encoded = json.dumps(payload, sort_keys=True, default=str).encode()
    return hashlib.sha256(encoded).hexdigest()


def _latest_audit_hash(db: Session, contract_id: str | None) -> str | None:
    stmt = select(AuditEvent.event_hash).order_by(AuditEvent.created_at.desc()).limit(1)
    if contract_id is not None:
        stmt = stmt.where(AuditEvent.contract_id == contract_id)
    else:
        stmt = stmt.where(AuditEvent.contract_id.is_(None))
    return db.scalar(stmt)


def record_access_event(
    db: Session,
    *,
    actor: str,
    resource_type: str,
    action: str,
    resource_id: str | None = None,
    ip: str | None = None,
) -> AccessEvent:
    event = AccessEvent(
        id=str(uuid4()),
        actor=actor,
        resource_type=resource_type,
        resource_id=resource_id,
        action=action,
        ip=ip,
        created_at=datetime.now(timezone.utc),
    )
    db.add(event)
    db.commit()
    db.refresh(event)
    return event


def record_audit_event(
    db: Session,
    *,
    actor: str,
    action: str,
    contract_id: str | None = None,
    payload: dict[str, Any] | None = None,
    prev_event_hash: str | None = None,
) -> AuditEvent:
    chain_prev = _latest_audit_hash(db, contract_id)
    if prev_event_hash is not None and prev_event_hash != chain_prev:
        raise AuditChainError("prev_event_hash does not match audit chain tip")

    prev_event_hash = chain_prev
    body = payload or {}
    event_hash = _hash_payload(
        {
            "actor": actor,
            "action": action,
            "contract_id": contract_id,
            "payload": body,
            "prev_event_hash": prev_event_hash,
        }
    )
    event = AuditEvent(
        id=str(uuid4()),
        contract_id=contract_id,
        actor=actor,
        action=action,
        payload_json=body,
        prev_event_hash=prev_event_hash,
        event_hash=event_hash,
        created_at=datetime.now(timezone.utc),
    )
    db.add(event)
    db.commit()
    db.refresh(event)
    return event
