import os

from app.domain.database import Base
from sqlalchemy.engine import URL
from sqlalchemy.ext.asyncio import create_async_engine


def load_db_config():
    db_user = os.getenv("AI_AGENT_DB_USER")
    db_pass = os.getenv("AI_AGENT_DB_PASSWORD")
    db_host = os.getenv("AI_AGENT_DB_HOST")
    db_port = os.getenv("AI_AGENT_DB_PORT", "5432")
    db_name = "ai_agent_db"

    return URL.create(
        drivername="postgresql+asyncpg",
        username=db_user,
        password=db_pass,
        host=db_host,
        port=int(db_port),
        database=db_name,
    )


DATABASE_URL = load_db_config()
engine = create_async_engine(DATABASE_URL, echo=True)


async def init_db():
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)

    print("database tables initialized")
