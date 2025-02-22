const { Kafka } = require("kafkajs");
const { exec } = require("child_process");

const kafka = new Kafka({ brokers: ["kafka:9092"] });
const consumer = kafka.consumer({ groupId: "executor-nodejs" });
const producer = kafka.producer();

const TASK_TOPIC = "code-execution-tasks-nodejs";
const RESULT_TOPIC = "code-execution-results";

(async () => {
    await consumer.connect();
    await producer.connect();
    await consumer.subscribe({ topic: TASK_TOPIC, fromBeginning: false });

    console.log("Node.js Executor listening for tasks...");

    await consumer.run({
        eachMessage: async ({ topic, partition, message }) => {
            const task = JSON.parse(message.value.toString());
            console.log(`Executing Node.js code: ${task.code}`);

            exec(`node -e "${task.code.replace(/"/g, '\\"')}"`, (error, stdout, stderr) => {
                const result = {
                    task_id: message.key.toString(),
                    output: stdout,
                    error: stderr || (error ? error.message : "")
                };

                producer.send({
                    topic: RESULT_TOPIC,
                    messages: [{ key: message.key.toString(), value: JSON.stringify(result) }]
                });

                console.log("Execution result for task ID", message.key.toString(), "sent to Kafka",);
            });
        },
    });
})();
