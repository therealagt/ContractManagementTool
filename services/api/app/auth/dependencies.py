from fastapi import Depends, Header, HTTPException, Request, status
from sqlalchemy.orm import Session

from app.auth.iap import IAP_HEADER, IapAuthError, IapUser, validate_iap_jwt
from app.db.session import get_db
from libs.common.audit.events import record_access_event


def get_iap_user(
    request: Request,
    x_goog_iap_jwt_assertion: str | None = Header(default=None, alias=IAP_HEADER),
) -> IapUser:
    token = x_goog_iap_jwt_assertion
    if not token:
        raise HTTPException(status.HTTP_401_UNAUTHORIZED, "Missing IAP JWT")

    try:
        return validate_iap_jwt(token)
    except IapAuthError as exc:
        raise HTTPException(status.HTTP_401_UNAUTHORIZED, str(exc)) from exc


def get_client_ip(request: Request) -> str | None:
    forwarded = request.headers.get("x-forwarded-for")
    if forwarded:
        return forwarded.split(",")[0].strip()
    if request.client:
        return request.client.host
    return None


class AccessLogger:
    def __init__(self, db: Session, user: IapUser, request: Request):
        self.db = db
        self.user = user
        self.ip = get_client_ip(request)

    def log(self, resource_type: str, action: str, resource_id: str | None = None) -> None:
        record_access_event(
            self.db,
            actor=self.user.email,
            resource_type=resource_type,
            action=action,
            resource_id=resource_id,
            ip=self.ip,
        )


def get_access_logger(
    request: Request,
    db: Session = Depends(get_db),
    user: IapUser = Depends(get_iap_user),
) -> AccessLogger:
    return AccessLogger(db, user, request)
