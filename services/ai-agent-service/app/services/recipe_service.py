import os
import uuid

import httpx
from app.domain.database import SavedWebRecipe
from app.types.request import SaveRecipeRequest
from sqlalchemy import select
from sqlalchemy.ext.asyncio import async_sessionmaker


class RecipeService:
    def __init__(self, session_factory: async_sessionmaker):
        self.session_factory = session_factory
        self.RECIPE_SERVICE_URL = os.getenv(
            "RECIPE_SERVICE_URL", "http://localhost:8083"
        )

    async def find_by_url(self, url: str):
        async with self.session_factory() as session:
            query = select(SavedWebRecipe).where(SavedWebRecipe.source_url == url)

            result = await session.execute(query)
            record = result.scalars().first()

            return record.recipe_id if record else None

    async def create_recipe_from_ai(
        self, recipe: SaveRecipeRequest, auth_header: str
    ) -> uuid.UUID:
        payload = {
            "title": recipe.title,
            "ingredients": recipe.ingredients,
            "instructions": recipe.instructions,
            "prep_time_min": recipe.total_time_minutes,
        }

        async with httpx.AsyncClient() as client:
            try:
                response = await client.post(
                    f"{self.RECIPE_SERVICE_URL}/recipes/",
                    json=payload,
                    headers={"Authorization": auth_header},
                    timeout=10.0,
                )
                response.raise_for_status()

                main_recipe_id = uuid.UUID(response.json()["id"])
            except httpx.HTTPStatusError as e:
                print(f"failed to push to recipe-service: {e.response.text}")

                raise

        async with self.session_factory() as session:
            try:
                new_mapping = SavedWebRecipe(
                    recipe_id=main_recipe_id, source_url=recipe.source_url
                )
                session.add(new_mapping)

                await session.commit()
            except Exception as e:
                print(f"WARNING: could not save URL mapping locally: {e}")

                await session.rollback()

        return main_recipe_id
