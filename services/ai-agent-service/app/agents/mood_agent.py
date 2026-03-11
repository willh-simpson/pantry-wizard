import json
import re
from typing import List

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


class WebRecipe(BaseModel):
    title: str = Field(description="recipe name")
    description: str = Field(description="brief recipe summary")
    source_url: str = Field(
        description='link to full recipe if available, otherwise "search snippet"'
    )
    key_ingredients: List[str] = Field(description="list of main ingredients mentioned")
    prep_time_minutes: int = Field(
        description="total time in minutes as integer. use 0 if unknown"
    )


class WebRecipeResponse(BaseModel):
    recipes: List[WebRecipe] = Field(description="list of extracted recipes")


class MoodAgent:
    def __init__(self):
        self.llm = OllamaLLM(model="llama3", temperature=0)
        self.query_parser = PydanticOutputParser(pydantic_object=SearchQuery)
        self.web_parser = PydanticOutputParser(pydantic_object=WebRecipeResponse)
        self.search_tool = DuckDuckGoSearchResults()

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
