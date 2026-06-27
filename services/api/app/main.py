from contextlib import asynccontextmanager

from fastapi import FastAPI

from app.config import settings
from app.db.session import run_migrations
from app.routes import health


@asynccontextmanager
async def lifespan(_: FastAPI):
    settings.validate_security()
    run_migrations()
    yield


app = FastAPI(title="Contract Management API", lifespan=lifespan)
app.include_router(health.router)
