from typing import List

from langchain_core.output_parsers import PydanticOutputParser
from langchain_core.prompts import PromptTemplate
from langchain_ollama import OllamaLLM
from pydantic import BaseModel, Field


class SearchQuery(BaseModel):
    semantic_query: str = Field(description="descriptive string for vector search")
    tags: List[str] = Field(description="specific ingredient or category tags")
    max_prep_time: int = Field(
        description="estimated maximum minutes based on how much time the user wants to spend cooking"
    )


class MoodAgent:
    def __init__(self):
        self.llm = OllamaLLM(model="llama3")
        self.parser = PydanticOutputParser(pydantic_object=SearchQuery)

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
                "format_instructions": self.parser.get_format_instructions()
            },
        )

        full_prompt = prompt.format(user_input=user_input)
        response = self.llm.invoke(full_prompt)

        try:
            return self.parser.parse(response)
        except Exception:
            return SearchQuery(semantic_query=user_input, tags=[], max_prep_time=60)
