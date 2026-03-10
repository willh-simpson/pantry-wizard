import httpx
from app.agents.mood_processor import MoodAgent

RESULTS_SCORE_THRESHOLD = 1.2

mood_agent = MoodAgent()


async def search_with_prompt(query: str, search_url: str):
    structured_query = mood_agent.process_mood(query)

    async with httpx.AsyncClient() as client:
        search_response = await client.post(
            search_url,
            params={"query": structured_query.semantic_query, "top_k": 5},
            timeout=10.0,
        )
        search_response.raise_for_status()

        local_results = search_response.json().get("results", [])

    # in FAISS L2 lower scores are better (0.0 is perfect match)
    # if the local scores are too high then move to web search
    web_results = None
    if not local_results or local_results[0]["score"] > RESULTS_SCORE_THRESHOLD:
        web_results = mood_agent.search_the_web(structured_query.semantic_query)

    return structured_query, local_results, web_results
