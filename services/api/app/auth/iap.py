from dataclasses import dataclass

from google.auth.transport import requests
from google.oauth2 import id_token

from app.config import settings

IAP_HEADER = "x-goog-iap-jwt-assertion"


@dataclass(frozen=True)
class IapUser:
    email: str
    sub: str


class IapAuthError(Exception):
    pass


def _email_domain(email: str) -> str:
    _, _, domain = email.rpartition("@")
    return domain.lower()


def validate_iap_jwt(token: str) -> IapUser:
    if settings.iap_jwt_validation_disabled:
        if settings.environment == "prod":
            raise IapAuthError("IAP JWT validation is disabled in prod")
        return IapUser(email="local-dev@internal", sub="local-dev")

    if not settings.iap_audience:
        raise IapAuthError("IAP audience not configured")

    try:
        info = id_token.verify_oauth2_token(
            token, requests.Request(), audience=settings.iap_audience
        )
    except ValueError as exc:
        raise IapAuthError("Invalid IAP JWT") from exc

    email = info.get("email")
    sub = info.get("sub")
    if not email or not sub:
        raise IapAuthError("IAP JWT missing identity claims")

    if not info.get("email_verified", False):
        raise IapAuthError("IAP JWT email is not verified")

    allowed_domains = settings.parsed_allowed_email_domains()
    if allowed_domains:
        domain = _email_domain(email)
        hosted_domain = info.get("hd")
        if domain not in allowed_domains and hosted_domain not in allowed_domains:
            raise IapAuthError("Email domain is not allowed")

    return IapUser(email=email.lower(), sub=sub)
