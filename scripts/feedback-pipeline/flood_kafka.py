import json
import random
import time
import uuid

from kafka import KafkaProducer

producer = KafkaProducer(
    bootstrap_servers=["localhost:9092"],
    value_serializer=lambda v: json.dumps(v).encode("utf-8"),
)

searches = [
    "spicy noodles",
    "vegan tacos",
    "quick breakfast",
    "chocolate cake",
    "keto dinner",
]

print("starting data flood...")
while True:
    data = {
        "userId": f"user_{str(uuid.uuid4())}",
        "action": "SEARCH",
        "targetId": None,
        "timestamp": int(time.time() * 1000),
        "metadata": {
            "q": random.choice(searches),
        },
    }

    producer.send("interactions.pantry.raw", data)
    print(f"sent {data} to producer")

    time.sleep(0.5)
