import httpx
from app.agents.mood_agent import MoodAgent
from app.types.faiss_scores import STRONG_SCORE_THRESHOLD

mood_agent = MoodAgent()


async def search_with_prompt(query: str, search_url: str):
    structured_query = mood_agent.process_mood(query)
    max_time = structured_query.max_prep_time

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
    web_results = []
    if not local_results or local_results[0]["score"] > STRONG_SCORE_THRESHOLD:
        raw_web_response = mood_agent.search_the_web(structured_query.semantic_query)
        web_results = mood_agent.summarize_web_results(query, raw_web_response)

        # only keep recipes either under max_time or time is unknown
        filtered_web_results = [
            r
            for r in web_results
            if r.prep_time_minutes <= max_time or r.prep_time_minutes == 0
        ]

    return structured_query, local_results, filtered_web_results
