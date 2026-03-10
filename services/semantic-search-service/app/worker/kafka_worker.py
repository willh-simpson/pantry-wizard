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


def print_consumer_message(event):
    # instruction field is very long. trim it to keep logs clean
    instructions = event["instructions"]
    trimmed_instructions = instructions[:10] + "..."

    print(
        "{'id': '%s', 'title': '%s', 'description': '%s', 'instructions':'%s'}"
        % (event["id"], event["title"], event["description"], trimmed_instructions)
    )


def kafka_worker(bootstrap_servers, topic_name):
    ensure_topic_exists(bootstrap_servers, topic_name)

    config = {
        "bootstrap.servers": bootstrap_servers,
        "group.id": "semantic-search-group",
        "auto.offset.reset": "earliest",
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
            print_consumer_message(event)

            if event["event_type"] in ["RECIPE_CREATED", "RECIPE_UPDATED"]:
                id = event["id"]
                title = event["title"]
                description = event["description"]
                instructions = event["instructions"]

                index_recipe(id, title, description, instructions)
    finally:
        consumer.close()
