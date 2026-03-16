import asyncio
import json
import re
from typing import List

from langchain_community.document_loaders import AsyncHtmlLoader, PlaywrightURLLoader
from langchain_community.document_transformers import BeautifulSoupTransformer
from langchain_community.tools import DuckDuckGoSearchResults
from langchain_core.output_parsers import PydanticOutputParser
from langchain_core.prompts import PromptTemplate
from langchain_ollama import OllamaLLM
from pydantic import BaseModel, Field


class SearchQuery(BaseModel):
    semantic_query: str = Field(description="descriptive string for vector search")
    tags: List[str] = Field(description="specific ingredient or category tags")
    max_prep_time: int = Field(
        description="estimated maximum minutes based on how much time the user wants to spend cooking. default to 60"
    )


class IngredientMatch(BaseModel):
    ingredient: str = Field(description="original ingredient name from the recipe")
    status: str = Field(description="one of: 'HAVE', 'MISSING', or 'SUBSTITUTE'")
    pantry_item: str = Field(
        description="the matching item in the user's pantry, or 'None'", default="None"
    )
    note: str = Field(
        description="reasoning for substitution or quantity notes", default=""
    )


class SoftPantryReport(BaseModel):
    missing_counts: List[int] = Field(
        description="the number of missing ingredients for each recipe in order"
    )


class PantryGapReport(BaseModel):
    matches: List[IngredientMatch] = Field(
        description="list of ingredients in recipe and their ingredient match report"
    )
    missing_count: int = Field(
        description="amount of ingredients in recipe not in user's pantry (substitutes do not count)"
    )
    can_make_now: bool = Field(
        description="True if 0 missing ingredients (substitutes count as having)",
        default=False,
    )


class WebRecipe(BaseModel):
    title: str = Field(description="recipe name")
    description: str = Field(description="brief recipe summary")
    source_url: str = Field(
        description='link to full recipe if available, otherwise "search snippet"'
    )
    key_ingredients: List[str] = Field(description="list of main ingredients mentioned")
    missing_ingredients_count: int = Field(
        description="amount of key ingredients from search result not in user's pantry",
        default=0,
    )
    prep_time_minutes: int = Field(
        description="total time in minutes as integer", default=0
    )


class WebRecipeResponse(BaseModel):
    recipes: List[WebRecipe] = Field(description="list of extracted recipes")


class FullRecipe(BaseModel):
    title: str = Field(description="recipe name")
    ingredients: List[str] = Field(
        description="list of recipe ingredients from provided recipe"
    )
    instructions: List[str] = Field(
        description="list of step-by-step instructions on how to make recipe"
    )
    servings: str = Field(
        description="amount of servings recipe makes", default="Not specified"
    )
    total_time_minutes: int = Field(
        description="total time in minutes as integer", default=0
    )


class MoodAgent:
    def __init__(self):
        self.llm = OllamaLLM(model="llama3", temperature=0)
        self.query_parser = PydanticOutputParser(pydantic_object=SearchQuery)
        self.web_parser = PydanticOutputParser(pydantic_object=WebRecipeResponse)
        self.deep_web_parser = PydanticOutputParser(pydantic_object=FullRecipe)
        self.pantry_gap_parser = PydanticOutputParser(pydantic_object=PantryGapReport)
        self.soft_pantry_gap_parser = PydanticOutputParser(
            pydantic_object=SoftPantryReport
        )
        self.search_tool = DuckDuckGoSearchResults()

    def _get_empty_recipe(self, url: str) -> FullRecipe:
        return FullRecipe(
            title="Recipe extraction failed",
            ingredients=["Could not extract ingredients"],
            instructions=["Visit source URL to view full instructions"],
            servings="Unknown",
            total_time_minutes=0,
        )

    # this will execute only if insufficient matches are found in local db
    def search_the_web(self, query: str):
        print(f"searching the web for: {query}")

        return self.search_tool.run(f"{query} recipe ingredients prep time")

    def process_mood(self, user_input: str) -> SearchQuery:
        prompt_text = """
            You are a culinary assistant for the 'Pantry Wizard' app.
            Translate the user's mood and request into a structured search query.

            User input: {user_input}

            {format_instructions}
        """

        prompt = PromptTemplate(
            template=prompt_text,
            input_variables=["user_input"],
            partial_variables={
                "format_instructions": self.query_parser.get_format_instructions()
            },
        )

        full_prompt = prompt.format(user_input=user_input)
        response = self.llm.invoke(full_prompt)

        try:
            return self.query_parser.parse(response)
        except Exception:
            return SearchQuery(semantic_query=user_input, tags=[], max_prep_time=60)

    def summarize_web_results(
        self, user_query: str, search_text: str
    ) -> List[WebRecipe]:
        if not search_text or len(search_text) < 20:
            print("WARNING: web search returned no meaningful text")

            return []

        prompt_text = """
            [INST]
            You are a JSON generator. Extract as many recipes as are provided from the search data provided.

            Your output must be a JSON object containing a "recipes" key.
            For every recipe, include every field in the schema.
            
            IMPORTANT:
            1. 'prep_time_minutes' must be a single integer represented as minutes. i.e. if the text says "1 hour", write 60.
            2. 'source_url' must be the actual link provided in the search metadata.
            3. 'key_ingredients' should include all ingredients mentioned in the result.
            4. If info is missing for a string field, use "Information not available".
            [/INST]

            User query: {user_query}
            Search Data: {search_text}

            {format_instructions}
        """

        prompt = PromptTemplate(
            template=prompt_text,
            input_variables=["user_query", "search_text"],
            partial_variables={
                "format_instructions": self.web_parser.get_format_instructions()
            },
        )
        full_prompt = prompt.format(user_query=user_query, search_text=search_text)

        print("DEBUG: sending request to llama 3 for synthesis...")

        response = self.llm.invoke(full_prompt)
        print(f"DEBUG: raw llm response: \n{response}")

        # llm will most likely have filler text in its response with the JSON list even if instructed not to
        json_match = re.search(r"(\{.*\})", response, re.DOTALL)
        clean_json = json_match.group(1) if json_match else response

        try:
            parsed_web_results = self.web_parser.parse(clean_json)

            return parsed_web_results.recipes
        except Exception as e:
            print(
                f"WARNING: json extraction failed: {e}. attempting manual parsing instead"
            )

            pass

        # attempting to manually parse in case wrapper fails
        try:
            list_match = re.search(r"(\[.*\])", response, re.DOTALL)

            if list_match:
                raw_list = json.loads(list_match.group(1))

                return [WebRecipe(**r) for r in raw_list]
        except Exception as e:
            print(f"WARNING: manual json parsing failed: {e}")

            return []

    # AsyncHmtlLoader may fail if site uses javascript to write information after page loads, e.g., tikok, pinterest
    # if docs[0].page_content is a bunch of <script> tags then use this
    async def deep_read_javascript_results(self, url: str):
        loader = PlaywrightURLLoader(
            urls=[url], remove_selectors=["header", "footer", "ad"]
        )
        docs = await loader.aload()

        return docs

    async def scrape_recipe_and_summarize(self, url: str) -> FullRecipe:
        print(f"performing deep-read at: {url}")

        # bypass basic bot detection by defining 'human' headers
        # more advanced bot detection on websites will stop me ¯\_(ツ)_/¯
        headers = {
            "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.37 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
            "Accept": "text/html.application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
            "Accept-Language": "en-US,en;q=0.9",
            "Referer": "https://www.google.com/",  # simulating request is coming from search engine. in theory this is already true given search_the_web()
        }

        try:
            loader = AsyncHtmlLoader([url], header_template=headers)
            docs = await asyncio.wait_for(asyncio.to_thread(loader.load), timeout=15)

            if not docs or not docs[0].page_content:
                raise ValueError("page loaded but content is empty")
        except asyncio.TimeoutError:
            print(f"ERROR: scraping '{url}' timed out")

            return self._get_empty_recipe(url)
        except Exception as e:
            print(f"ERROR: scraping failed for '{url}': {e}")

            pass

        # less than 500 characters is suspiciously short or could just be a loading screen
        if not docs or len(docs[0].page_content) < 500:
            print(
                f"DEBUG: content too thin or missing. attempting scrape with Playwright for '{url}'"
            )

            try:
                docs = await self.deep_read_javascript_results(url)
            except Exception as e:
                print(f"ERROR: playwright scrape failed: {e}")

                return self._get_empty_recipe(url)

        bs_transformer = BeautifulSoupTransformer()
        docs_transformed = bs_transformer.transform_documents(
            docs, tags_to_extract=["p", "li", "div", "span"]
        )

        # taking first 4000 tokens to avoid hitting context limits
        page_content = docs_transformed[0].page_content[:12000]
        prompt_text = """
            [INST]
            You are a culinary assistant for the 'Pantry Wizard' app.
            Your job is to extract the full recipe from the below webpage content into a clean JSON format.

            "total_time_minutes" JSON field is the total time in minutes to make the recipe. 
            i.e. if the recipe says "1 hour", this field should be set to 60.

            Ignore all ads, related posts, and website navigation.
            [/INST]

            Webpage content:
            {page_content}

            {format_instructions}
        """

        prompt = PromptTemplate(
            template=prompt_text,
            input_variables=["page_content"],
            partial_variables={
                "format_instructions": self.deep_web_parser.get_format_instructions(),
            },
        )

        full_prompt = prompt.format(page_content=page_content)
        response = self.llm.invoke(full_prompt)

        json_match = re.search(r"(\{.*\})", response, re.DOTALL)
        clean_json = json_match.group(1) if json_match else response

        return self.deep_web_parser.parse(clean_json)

    async def soft_analyze_pantry_gap(
        self, recipes: List[WebRecipe], user_pantry: List[int]
    ):
        if not recipes:
            return []

        recipes_str = "\n".join(
            [
                f"{i + 1}. {r.title}: {', '.join(r.key_ingredients)}"
                for i, r in enumerate(recipes)
            ]
        )

        prompt_text = """
            [INST]
            Compare the key ingredients of the provided recipes against the User pantry.
            Return only the number of MISSING ingredients for each recipe.

            RULES:
            - If the user pantry has the ingredient or a valid substitute, it is not missing.
            - A valid substitute is a common, flavor-appropriate replacement that will not affect the recipe (e.g., 'onion' for 'shallot').
            - Only return the interger count of missing ingredients per recipe.
            [/INST]

            User pantry: {user_pantry}

            Recipes: {recipes}

            {format_instructions}
        """

        prompt = PromptTemplate(
            template=prompt_text,
            input_variables=["user_pantry", "recipes"],
            partial_variables={
                "format_instructions": self.soft_pantry_gap_parser.get_format_instructions()
            },
        )

        full_prompt = prompt.format(
            user_pantry=", ".join(user_pantry), recipes=recipes_str
        )

        try:
            response = self.llm.invoke(full_prompt)

            json_match = re.search(r"(\{.*\})", response, re.DOTALL)
            result = self.soft_pantry_gap_parser.parse(
                json_match.group(1) if json_match else response
            )

            if len(result.missing_counts) != len(recipes):
                return [0] * len(recipes)

            return result.missing_counts
        except Exception as e:
            print(f"soft pantry check failed: {e}")

            return [0] * len(recipes)

    async def hard_analyze_pantry_gap(
        self, recipe_ingredients: List[str], user_pantry: List[str]
    ) -> PantryGapReport:
        prompt_text = """
            [INST]
            You are an expert chef and inventory manager.
            Compare the "Recipe ingredients" against the "User pantry".

            LOGIC:
            - MATCH: if the user pantry has the item (e.g., 'olive oil' matches '1 tbsp extra virgin olive oil').
            - SUBSTITUTE: if the pantry item is a common, flavor-appropriate replacement (e.g., 'onion' for 'shallot').
            - MISSING: if there is no reasonable match.
            [/INST]

            User pantry: {user_pantry}
            Recipe ingredients: {recipe_ingredients}

            {format_instructions}
        """

        prompt = PromptTemplate(
            template=prompt_text,
            input_variables=["user_pantry", "recipe_ingredients"],
            partial_variables={
                "format_instructions": self.pantry_gap_parser.get_format_instructions()
            },
        )

        full_prompt = prompt.format(
            user_pantry=", ".join(user_pantry),
            recipe_ingredients=", ".join(recipe_ingredients),
        )

        print("DEBUG: analyzing pantry gap...")
        response = self.llm.invoke(full_prompt)

        json_match = re.search(r"(\{.*\})", response, re.DOTALL)
        clean_json = json_match.group(1) if json_match else response

        return self.pantry_gap_parser.parse(clean_json)
