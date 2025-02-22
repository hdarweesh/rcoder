FROM gcc:latest

WORKDIR /app

# Install dependencies for Kafka C++ client
RUN apt-get update && apt-get install -y librdkafka-dev

COPY . .

# Compile the executor
RUN g++ -o executor executor.cpp -lrdkafka -lpthread -std=c++11

CMD ["./executor"]
