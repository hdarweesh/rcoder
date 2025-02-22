import json
import time
from kafka import KafkaConsumer, KafkaProducer
from subprocess import Popen, PIPE

KAFKA_BROKER = "kafka:9092"
TASK_TOPIC = "code-execution-tasks-python"
RESULT_TOPIC = "code-execution-results"

# FIX: Retry connecting to Kafka until available
def wait_for_kafka():
    for i in range(10):
        try:
            consumer = KafkaConsumer(TASK_TOPIC, bootstrap_servers=KAFKA_BROKER)
            consumer.close()
            print("Connected to Kafka!")
            return
        except Exception as e:
            print(f"⏳ Waiting for Kafka... Retry {i+1}/10")
            time.sleep(5)
    print("Kafka is not reachable. Exiting.")
    exit(1)

# Ensure Kafka is available before starting the consumer
wait_for_kafka()

consumer = KafkaConsumer(
    TASK_TOPIC,
    bootstrap_servers=KAFKA_BROKER,
    auto_offset_reset="earliest",
    enable_auto_commit=True,
    group_id="executor-python",
)

producer = KafkaProducer(bootstrap_servers=KAFKA_BROKER)

def execute_python_code(code):
    try:
        process = Popen(["python3", "-c", code], stdout=PIPE, stderr=PIPE)
        stdout, stderr = process.communicate(timeout=5)
        return stdout.decode(), stderr.decode()
    except Exception as e:
        return "", str(e)

for message in consumer:
    task = json.loads(message.value)
    print(f"Executing Python code: {task['code']}")

    stdout, stderr = execute_python_code(task["code"])

    response = {
        "task_id": message.key.decode(),
        "output": stdout.strip(),
        "error": stderr.strip(),
    }

    producer.send(RESULT_TOPIC, json.dumps(response).encode("utf-8"))
    print("Execution result sent to Kafka.")
