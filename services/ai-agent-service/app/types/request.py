from typing import List

from pydantic import BaseModel


class ChatRequest(BaseModel):
    message: str


class ExtractionRequest(BaseModel):
    url: str


class SaveRecipeRequest(BaseModel):
    title: str
    source_url: str
    ingredients: List[str]
    instructions: List[str]
    servings: str
    total_time_minutes: int
