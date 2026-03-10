import os

import httpx
from app.agents.mood_processor import MoodAgent
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
        structured_query = mood_agent.process_mood(request.message)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"llm processing error: {str(e)}")

    async with httpx.AsyncClient() as client:
        try:
            search_response = await client.post(
                f"{SEMANTIC_SEARCH_URL}/search",
                params={"query": structured_query.semantic_query, "top_k": 5},
                timeout=10.0,
            )

            search_response.raise_for_status()
            search_results = search_response.json()
        except httpx.HTTPError:
            return {
                "agent_analysis": structured_query,
                "results": [],
                "warning": "sematnic search service unreachable.",
            }

    return {
        "analysis": {
            "mood": request.message,
            "translated_query": structured_query.semantic_query,
            "detected_tags": structured_query.tags,
            "prep_limit": structured_query.max_prep_time,
        },
        "matches": search_results.get("results", []),
    }
