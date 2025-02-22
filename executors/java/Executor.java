import org.apache.kafka.clients.consumer.*;
import org.apache.kafka.clients.producer.*;

import java.io.*;
import java.util.Collections;
import java.util.Properties;

public class Executor {
    private static final String KAFKA_BROKER = "kafka:9092";
    private static final String TASK_TOPIC = "code-execution-tasks-java";
    private static final String RESULT_TOPIC = "code-execution-results";

    public static void main(String[] args) {
        Properties props = new Properties();
        props.put(ConsumerConfig.BOOTSTRAP_SERVERS_CONFIG, KAFKA_BROKER);
        props.put(ConsumerConfig.GROUP_ID_CONFIG, "executor-java");
        props.put(ConsumerConfig.KEY_DESERIALIZER_CLASS_CONFIG, "org.apache.kafka.common.serialization.StringDeserializer");
        props.put(ConsumerConfig.VALUE_DESERIALIZER_CLASS_CONFIG, "org.apache.kafka.common.serialization.StringDeserializer");

        KafkaConsumer<String, String> consumer = new KafkaConsumer<>(props);
        consumer.subscribe(Collections.singletonList(TASK_TOPIC));

        KafkaProducer<String, String> producer = new KafkaProducer<>(props);

        while (true) {
            for (ConsumerRecord<String, String> record : consumer.poll(1000)) {
                String code = record.value();
                String output = executeJavaCode(code);

                ProducerRecord<String, String> result = new ProducerRecord<>(RESULT_TOPIC, record.key(), output);
                producer.send(result);
            }
        }
    }

    private static String executeJavaCode(String code) {
        try {
            Process process = new ProcessBuilder("java", "-cp", ".", code).start();
            BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()));
            return reader.readLine();
        } catch (IOException e) {
            return e.getMessage();
        }
    }
}
