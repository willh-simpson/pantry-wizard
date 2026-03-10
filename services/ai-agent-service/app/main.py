import os

import httpx
from app.agents.mood_processor import MoodAgent
from app.api.prompt_processor import search_with_prompt
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

app = FastAPI(title="Pantry Wizard AI Agent")
mood_agent = MoodAgent()

SEMANTIC_SEARCH_URL = os.getenv("SEMANTIC_SEARCH_SERVICE_URL", "localhost:8000")
print(f"connected to semantic-search-service via: {SEMANTIC_SEARCH_URL}")


class ChatRequest(BaseModel):
    message: str


@app.post("/ai/predict-mood")
async def predict_mood_and_search(request: ChatRequest):
    try:
        structured_query, local_results, web_results = await search_with_prompt(
            request.message, f"{SEMANTIC_SEARCH_URL}/search"
        )
    except httpx.HTTPError:
        return {
            "agent_analysis": structured_query,
            "results": [],
            "warning": "semantic search service unreachable.",
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"llm processing error: {str(e)}")

    return {
        "analysis": structured_query,
        "local_matches": local_results,
        "web_matches": web_results,
    }
