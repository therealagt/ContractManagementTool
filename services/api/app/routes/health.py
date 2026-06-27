from fastapi import APIRouter, Depends

from app.auth.dependencies import AccessLogger, get_access_logger, get_iap_user
from app.auth.iap import IapUser
from app.auth.roles import Role, require_roles, roles_for_user
from app.config import settings

router = APIRouter(tags=["health"])


@router.get("/health")
def health() -> dict[str, str]:
    # Unauthenticated: Cloud Run startup/liveness probes bypass IAP/LB.
    return {"status": "ok"}


@router.get("/status")
def status(
    user: IapUser = Depends(get_iap_user),
    access: AccessLogger = Depends(get_access_logger),
) -> dict[str, str | list[str]]:
    access.log("api", "status")
    return {
        "status": "ok",
        "environment": settings.environment,
        "actor": user.email,
        "roles": sorted(role.value for role in roles_for_user(user)),
    }


@router.get("/status/admin")
def status_admin(
    user: IapUser = Depends(require_roles(Role.ADMIN)),
    access: AccessLogger = Depends(get_access_logger),
) -> dict[str, str]:
    access.log("api", "status_admin")
    return {"status": "ok", "actor": user.email}
