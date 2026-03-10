import json
import os

import faiss
import numpy as np
from sentence_transformers import SentenceTransformer

model = SentenceTransformer("all-MiniLM-L6-v2")

index_file = "recipes.index"
map_file = "id_mapping.json"
dimension = 384  # dimension for all-MiniLM-L6-v2

if os.path.exists(index_file):
    index = faiss.read_index(index_file)

    with open(map_file, "r") as f:
        id_map = json.load(f)  # { "int_id": "{uuid}" }
else:
    index = faiss.IndexIDMap(faiss.IndexFlatL2(dimension))
    id_map = {}


def index_recipe(id, title, description, instructions):
    # this creates a searchable blob that the ml worker can execute
    text_to_embed = f"{title}. {description}. {instructions}"
    vector = model.encode([text_to_embed]).astype("float32")

    # faiss uses int ids, but sql schema uses uuid
    # this maps the given uuid to an incremental integer
    internal_id = len(id_map)
    index.add_with_ids(vector, np.array([internal_id]))
    id_map[str(internal_id)] = id

    faiss.write_index(index, index_file)
    with open(map_file, "w") as f:
        json.dump(id_map, f)

    print(f"indexed recipe: {title}")


def faiss_search(query: str, top_k: int = 5):
    query_vector = model.encode([query]).astype("float32")

    distances, internal_ids = index.search(query_vector, top_k)

    # faiss int ids need to be converted back to uuids before returning response
    results = []
    for i, idx in enumerate(internal_ids[0]):
        if str(idx) in id_map:
            results.append(
                {"recipe_id": id_map[str(idx)], "score": float(distances[0][i])}
            )

    return query, results
