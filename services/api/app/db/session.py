from collections.abc import Generator
from datetime import datetime, timezone

from sqlalchemy import create_engine, text
from sqlalchemy.engine import Engine
from sqlalchemy.orm import Session, sessionmaker

from app.config import settings
from app.db.models import Base

_engine: Engine | None = None
_SessionLocal: sessionmaker[Session] | None = None

MIGRATION_SQL = """
CREATE TABLE IF NOT EXISTS access_events (
    id VARCHAR(36) PRIMARY KEY,
    actor VARCHAR(320) NOT NULL,
    resource_type VARCHAR(64) NOT NULL,
    resource_id VARCHAR(64),
    action VARCHAR(64) NOT NULL,
    ip VARCHAR(45),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_events (
    id VARCHAR(36) PRIMARY KEY,
    contract_id VARCHAR(64),
    actor VARCHAR(320) NOT NULL,
    action VARCHAR(64) NOT NULL,
    payload_json JSONB,
    prev_event_hash VARCHAR(64),
    event_hash VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(32) PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_access_events_actor ON access_events(actor);
CREATE INDEX IF NOT EXISTS idx_audit_events_contract ON audit_events(contract_id);
"""


def get_engine() -> Engine:
    global _engine
    if _engine is None:
        connect_args = {}
        if settings.sqlalchemy_url.startswith("sqlite"):
            connect_args = {"check_same_thread": False}
        _engine = create_engine(settings.sqlalchemy_url, connect_args=connect_args, pool_pre_ping=True)
    return _engine


def get_session_factory() -> sessionmaker[Session]:
    global _SessionLocal
    if _SessionLocal is None:
        _SessionLocal = sessionmaker(bind=get_engine(), autocommit=False, autoflush=False)
    return _SessionLocal


def get_db() -> Generator[Session, None, None]:
    db = get_session_factory()()
    try:
        yield db
    finally:
        db.close()


def run_migrations() -> None:
    engine = get_engine()
    if settings.sqlalchemy_url.startswith("sqlite"):
        Base.metadata.create_all(bind=engine)
        return

    with engine.begin() as conn:
        for statement in MIGRATION_SQL.split(";"):
            sql = statement.strip()
            if sql:
                conn.execute(text(sql))
        conn.execute(
            text(
                "INSERT INTO schema_migrations (version, applied_at) "
                "VALUES (:version, :applied_at) ON CONFLICT DO NOTHING"
            ),
            {"version": "001_initial", "applied_at": datetime.now(timezone.utc)},
        )
