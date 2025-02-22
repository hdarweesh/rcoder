# Use official Python image
FROM python:3.11-alpine

# Set working directory
WORKDIR /app

# Copy dependencies first (to leverage caching)
COPY requirements.txt ./

# Install dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy the rest of the application
COPY . .

# Set entrypoint command
CMD ["sh", "-c", "sleep 5 && python3 executor.py"]