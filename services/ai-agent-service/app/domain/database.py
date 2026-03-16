from sqlalchemy import Column, String
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()


class SavedWebRecipe(Base):
    __tablename__ = "saved_web_links"

    recipe_id = Column(UUID(as_uuid=True), primary_key=True)
    source_url = Column(String, unique=True, nullable=False, index=True)
