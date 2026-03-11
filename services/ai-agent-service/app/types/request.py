from pydantic import BaseModel


class ChatRequest(BaseModel):
    message: str


class ExtractionRequest(BaseModel):
    url: str
