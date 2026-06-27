from datetime import datetime

from sqlalchemy import DateTime, String
from sqlalchemy.orm import Mapped, mapped_column

from libs.common.db.models import AccessEvent, AuditEvent, Base

__all__ = ["AccessEvent", "AuditEvent", "Base", "SchemaMigration"]


class SchemaMigration(Base):
    __tablename__ = "schema_migrations"

    version: Mapped[str] = mapped_column(String(32), primary_key=True)
    applied_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), nullable=False)
