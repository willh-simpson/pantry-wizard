import json
import random
import time
import uuid

from kafka import KafkaProducer

producer = KafkaProducer(
    bootstrap_servers=["localhost:9092"],
    value_serializer=lambda v: json.dumps(v).encode("utf-8"),
)

personas = [
    ("Bouncer", 0.6, 1, 1, 0),  # 60% search 1x then leave
    ("Researcher", 0.3, 3, 6, 5),  # 30% search 3-6x with 5sec gaps
    ("Chef", 0.1, 10, 20, 2),  # 10% search 10-20x rapidly
]

search_terms = [
    "keto dinner",
    "spicy noodles",
    "vegan tacos",
    "quick breakfast",
    "chocolate cake",
]

searches = [
    "spicy noodles",
    "vegan tacos",
    "quick breakfast",
    "chocolate cake",
    "keto dinner",
]


def send_interaction(user_id, action, query=None, target_id=None):
    payload = {
        "userId": user_id,
        "action": action,
        "targetId": target_id,
        "timestamp": int(time.time() * 1000),
        "metadata": {
            "q": query,
        },
    }

    producer.send("interactions.pantry.raw", payload)
    print(f"sent: {user_id} -> {query}")


def simulate_session():
    name, weight, min_s, max_s, gap = random.choices(
        personas, weights=[p[1] for p in personas]
    )[0]

    user_id = f"user_{uuid.uuid4().hex[:8]}"
    num_searches = random.randint(min_s, max_s)

    print(f"--- starting {name} session: {user_id} ({num_searches} searches) ---")

    for _ in range(num_searches):
        query = random.choice(search_terms)
        send_interaction(user_id, "SEARCH", query=query)

        if random.random() > 0.4:  # 60% chance of clicking on recipe
            time.sleep(random.uniform(1, 3))

            recipe_id = f"recipe_{random.randint(1, 500)}"
            send_interaction(user_id, "CLICK", target_id=recipe_id)

            if (
                random.random() > 0.7
            ):  # 30% chance of saving recipe after clicking (can't save without viewing recipe)
                time.sleep(random.uniform(1, 2))

                send_interaction(user_id, "SAVE", target_id=recipe_id)

    producer.flush()


if __name__ == "__main__":
    try:
        while True:
            simulate_session()

            time.sleep(random.randint(2, 10))
    except KeyboardInterrupt:
        print("simulation stopped")
