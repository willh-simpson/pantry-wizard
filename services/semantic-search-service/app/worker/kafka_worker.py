import json

from app.api.search import index_recipe
from confluent_kafka import Consumer
from confluent_kafka.admin import AdminClient, NewTopic


def ensure_topic_exists(bootstrap_servers, topic_name):
    admin_client = AdminClient({"bootstrap.servers": bootstrap_servers})
    new_topics = [NewTopic(topic_name, num_partitions=1, replication_factor=1)]

    fs = admin_client.create_topics(
        new_topics
    )  # if the topic already exists this should fail gracefully
    for topic, f in fs.items():
        try:
            f.result()

            print(f"topic {topic} created or verified")
        except Exception as e:
            print(f"topic {topic} check: {e}")


def kafka_worker(bootstrap_servers, topic_name):
    ensure_topic_exists(bootstrap_servers, topic_name)

    config = {
        "bootstrap.servers": bootstrap_servers,
        "group.id": "semantic-search-group",
        "auto.offset-reset": "earliest",
        "enable.auto.commit": True,
    }

    consumer = Consumer(config)
    consumer.subscribe([topic_name])

    print(f"kafka worker started. listening for {topic_name}...")

    try:
        while True:
            msg = consumer.poll(1.0)
            if msg is None:
                continue
            if msg.error():
                print(f"consumer error: {msg.error()}")

                continue

            event = json.loads(msg.value().decode("utf-8"))
            print(f"received message: {event}")

            if event["event_type"] in ["RECIPE_CREATED", "RECIPE_UPDATED"]:
                id = event["id"]
                title = event["title"]
                description = event["recipe_description"]
                instructions = event["recipe_instructions"]

                index_recipe(id, title, description, instructions)
    finally:
        consumer.close()
