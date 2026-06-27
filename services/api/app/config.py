from pydantic_settings import BaseSettings, SettingsConfigDict


class SecurityConfigError(RuntimeError):
    pass


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

    environment: str = "dev"
    gcp_project_id: str = ""
    gcp_region: str = "europe-west3"

    cloud_sql_connection_name: str = ""
    db_name: str = "contracts"
    db_user: str = "contract_api"
    db_password: str = ""
    database_url: str = ""

    gcs_staging_bucket: str = ""
    gcs_archive_bucket: str = ""

    iap_audience: str = ""
    iap_jwt_validation_disabled: bool = False
    allowed_email_domains: str = ""

    auth_uploader_emails: str = ""
    auth_reviewer_emails: str = ""
    auth_auditor_emails: str = ""
    auth_admin_emails: str = ""

    @property
    def sqlalchemy_url(self) -> str:
        if self.database_url:
            return self.database_url
        if not self.cloud_sql_connection_name:
            return "sqlite:///./local.db"
        socket = f"/cloudsql/{self.cloud_sql_connection_name}"
        return (
            f"postgresql+pg8000://{self.db_user}:{self.db_password}@/"
            f"{self.db_name}?unix_sock={socket}/.s.PGSQL.5432"
        )

    def parsed_allowed_email_domains(self) -> list[str]:
        return [d.strip().lower() for d in self.allowed_email_domains.split(",") if d.strip()]

    def parsed_email_set(self, raw: str) -> frozenset[str]:
        return frozenset(email.strip().lower() for email in raw.split(",") if email.strip())

    @property
    def uploader_emails(self) -> frozenset[str]:
        return self.parsed_email_set(self.auth_uploader_emails)

    @property
    def reviewer_emails(self) -> frozenset[str]:
        return self.parsed_email_set(self.auth_reviewer_emails)

    @property
    def auditor_emails(self) -> frozenset[str]:
        return self.parsed_email_set(self.auth_auditor_emails)

    @property
    def admin_emails(self) -> frozenset[str]:
        return self.parsed_email_set(self.auth_admin_emails)

    def validate_security(self) -> None:
        if self.environment == "prod":
            if self.iap_jwt_validation_disabled:
                raise SecurityConfigError("IAP JWT validation cannot be disabled in prod")
            if not self.iap_audience:
                raise SecurityConfigError("IAP audience is required in prod")
            if not self.parsed_allowed_email_domains():
                raise SecurityConfigError("allowed_email_domains is required in prod")

        if not self.iap_jwt_validation_disabled and not self.iap_audience:
            raise SecurityConfigError("IAP audience is required when JWT validation is enabled")


settings = Settings()
