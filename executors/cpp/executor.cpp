#include <iostream>
#include <fstream>
#include <sstream>
#include <cstdlib>
#include <string>
#include <vector>
#include <chrono>
#include <thread>
#include <librdkafka/rdkafkacpp.h>

#define KAFKA_BROKER "kafka:9092"
#define TASK_TOPIC "code-execution-tasks-cpp"
#define RESULT_TOPIC "code-execution-results"
#define TEMP_FILENAME "/app/code.cpp"
#define EXECUTABLE_NAME "/app/code.out"

std::string executeCppCode(const std::string& code) {
    // Write code to a file
    std::ofstream file(TEMP_FILENAME);
    if (!file) return "Error: Could not create C++ file.";
    file << code;
    file.close();

    // Compile the C++ file
    std::string compileCommand = "g++ " + std::string(TEMP_FILENAME) + " -o " + std::string(EXECUTABLE_NAME) + " 2>&1";
    FILE* compileOutput = popen(compileCommand.c_str(), "r");
    if (!compileOutput) return "Error: Compilation failed.";
    
    char buffer[128];
    std::ostringstream compileResult;
    while (fgets(buffer, sizeof(buffer), compileOutput) != nullptr) {
        compileResult << buffer;
    }
    pclose(compileOutput);

    if (!compileResult.str().empty()) return "Compilation Error: " + compileResult.str();

    // Run the compiled executable
    std::string executeCommand = std::string(EXECUTABLE_NAME) + " 2>&1";
    FILE* runOutput = popen(executeCommand.c_str(), "r");
    if (!runOutput) return "Error: Execution failed.";

    std::ostringstream output;
    while (fgets(buffer, sizeof(buffer), runOutput) != nullptr) {
        output << buffer;
    }
    pclose(runOutput);

    return output.str();
}

int main() {
    // Kafka Configuration
    std::string errstr;
    RdKafka::Conf* conf = RdKafka::Conf::create(RdKafka::Conf::CONF_GLOBAL);
    conf->set("bootstrap.servers", KAFKA_BROKER, errstr);
    
    RdKafka::Consumer* consumer = RdKafka::Consumer::create(conf, errstr);
    RdKafka::Topic* topic = RdKafka::Topic::create(consumer, TASK_TOPIC, nullptr, errstr);
    consumer->start(topic, 0, RdKafka::Topic::OFFSET_END);

    RdKafka::Producer* producer = RdKafka::Producer::create(conf, errstr);
    
    std::cout << "C++ Executor is running..." << std::endl;

    while (true) {
        RdKafka::Message* message = consumer->consume(topic, 0, 1000);
        if (message->err() == RdKafka::ERR_NO_ERROR) {
            std::string code(static_cast<char*>(message->payload()), message->len());
            std::string output = executeCppCode(code);

            RdKafka::Message* resultMessage = new RdKafka::Message();
            resultMessage->set_payload(output);
            producer->produce(RESULT_TOPIC, 0, output.size(), resultMessage->payload(), nullptr);
        }
        delete message;
        std::this_thread::sleep_for(std::chrono::milliseconds(500));
    }

    consumer->stop(topic, 0);
    delete consumer;
    delete producer;
    delete conf;

    return 0;
}
