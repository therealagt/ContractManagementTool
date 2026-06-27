from enum import StrEnum

from fastapi import Depends, HTTPException, status

from app.auth.dependencies import get_iap_user
from app.auth.iap import IapUser
from app.config import settings


class Role(StrEnum):
    UPLOADER = "uploader"
    REVIEWER = "reviewer"
    AUDITOR = "auditor"
    ADMIN = "admin"


_ALL_ROLES = frozenset(Role)


def roles_for_user(user: IapUser) -> frozenset[Role]:
    email = user.email.lower()
    roles: set[Role] = set()

    if email in settings.admin_emails:
        return _ALL_ROLES
    if email in settings.uploader_emails:
        roles.add(Role.UPLOADER)
    if email in settings.reviewer_emails:
        roles.add(Role.REVIEWER)
    if email in settings.auditor_emails:
        roles.add(Role.AUDITOR)

    return frozenset(roles)


def require_roles(*required: Role):
    def dependency(user: IapUser = Depends(get_iap_user)) -> IapUser:
        granted = roles_for_user(user)
        if not any(role in granted for role in required):
            raise HTTPException(
                status.HTTP_403_FORBIDDEN,
                "Insufficient permissions for this action",
            )
        return user

    return dependency


def assert_separation_of_duty(*, actor: str, other: str, action: str) -> None:
    if actor.lower() == other.lower():
        raise HTTPException(
            status.HTTP_403_FORBIDDEN,
            f"Separation of duties violation: same actor cannot {action}",
        )
