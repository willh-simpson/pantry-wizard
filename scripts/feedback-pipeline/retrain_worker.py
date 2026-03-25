import json
import os
import time

import torch
import torch.nn as nn
from kafka import KafkaConsumer

BOOTSTRAP_SERVERS = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
COMMAND_TOPIC = "pantry.retrian.commands"
GROUP_ID = "ml-orchestrator-group"

consumer = KafkaConsumer(
    COMMAND_TOPIC,
    bootstrap_servers=BOOTSTRAP_SERVERS,
    group_id=GROUP_ID,
    auto_offset_reset="latest",
    value_deserializer=lambda v: json.loads(v.decode("utf-8")),
)


class PantryRanker(nn.Module):
    def __init__(self):
        super().__init__()

        self.layers = nn.Sequential(
            nn.Linear(128, 64),  # 128-dim query embedding
            nn.ReLU(),
            nn.Linear(64, 1),
            nn.Sigmoid(),  # output probability between 0 and 1
        )

    def forward(self, x):
        return self.layers(x)


def train_pytorch(features, labels):
    model = PantryRanker()

    optimizer = torch.optim.Adam(model.parameters(), lr=0.001)
    criterion = nn.BCELoss()

    for epoch in range(10):
        optimizer.zero_grad()
        outputs = model(features)

        loss = criterion(outputs, labels)
        loss.backward()

        optimizer.step()

    torch.save(model.state_dict(), "pantry_ranker_v2.pth")


def update_model_metadata(version, metrics):
    metadata_path = "/opt/pantry/metadata/model_info.json"
    data = {
        "current_version": version,
        "last_trained": int(time.time()),
        "performance_at_train": metrics,
    }

    with open(metadata_path, "w") as f:
        json.dump(data, f)

    print(f"[METADATA] Updated {metadata_path} with version {version}")


def train_model(model_id, reason, samples):
    """
    Simulates the model training process.
    """

    print(f"[BOOTING] Starting retraining for: {model_id}")
    print(f"[REASON] {reason}")
    print(f"[DATASET] Using {samples} new interaction samples")

    for i in range(1, 6):
        time.sleep(1)

        progress = i * 20
        print(f"... Training progress: {progress}%")

    new_version = f"v{int(time.time())}"
    print(f"[SUCCESS Model {model_id} updated to version {new_version}")

    update_model_metadata(new_version, samples)

    return new_version


if __name__ == "__main__":
    print(f"Listening for retrain commands on topic {COMMAND_TOPIC}...")

    try:
        for message in consumer:
            command = message.value

            model_id = command.get("modelId", "unknown-model")
            reason = command.get("reason", "Scheduled maintenance")
            samples = command.get("sampleCount", 0)

            train_model(model_id, reason, samples)
    except KeyboardInterrupt:
        print("Worker shutting down...")
    finally:
        consumer.close()
