import os
import threading

from app.api.search import faiss_search
from app.worker.kafka_worker import kafka_worker
from fastapi import FastAPI

app = FastAPI()


def start_kafka_worker():
    kafka_bootstrap_servers = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")

    kafka_worker(kafka_bootstrap_servers, "recipe-events")


threading.Thread(target=start_kafka_worker, daemon=True).start()


@app.post("/search")
async def search(query: str, top_k: int = 5):
    search_query, results = faiss_search(query, top_k)

    return {
        "query": search_query,
        "results": results,
    }
