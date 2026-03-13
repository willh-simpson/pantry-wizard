import os
from contextlib import asynccontextmanager

import httpx
from app.domain import database_config
from app.services import search_service
from app.services.recipe_service import RecipeService
from app.types.request import ChatRequest, ExtractionRequest, SaveRecipeRequest
from fastapi import FastAPI, Header, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker


@asynccontextmanager
async def lifespan(app: FastAPI):
    await database_config.init_db()

    yield


app = FastAPI(title="Pantry Wizard AI Agent", lifespan=lifespan)

AsyncSessionLocal = async_sessionmaker(
    bind=database_config.engine, class_=AsyncSession, expire_on_commit=False
)
recipe_service = RecipeService(session_factory=AsyncSessionLocal)

SEMANTIC_SEARCH_URL = os.getenv("SEMANTIC_SEARCH_SERVICE_URL", "localhost:8000")
print(f"connected to semantic-search-service via: {SEMANTIC_SEARCH_URL}")


@app.post("/ai/predict-mood")
async def predict_mood_and_search(request: ChatRequest):
    try:
        (
            structured_query,
            local_results,
            web_results,
        ) = await search_service.search_with_prompt(
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


@app.post("/ai/extract-full-recipe")
async def extract_full_recipe(request: ExtractionRequest):
    try:
        return await search_service.deep_read_web_recipe(request.url)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"failed to read recipe: {str(e)}")


@app.post("/ai/save-web-recipe")
async def save_web_recipe(recipe: SaveRecipeRequest, authorization: str = Header(None)):
    try:
        existing_recipe = await recipe_service.find_by_url(recipe.source_url)

        if existing_recipe:
            return {
                "recipe_id": existing_recipe.id,
                "status": "already_exists",
            }

        new_recipe = await recipe_service.create_recipe_from_ai(recipe, authorization)

        return {
            "recipe_id": new_recipe.id,
            "status": "saved",
            "message": f"successfully saved {recipe.title} to library",
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
