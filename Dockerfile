# Use an official Python base image with build tools
FROM python:3.9-slim

# Install system dependencies for building eflomal
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    git \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Download fastalign deb and install it
RUN wget -c https://github.com/Unbabel/fast-align-deb/raw/refs/heads/master/fast-align_2016.05.31-1_amd64.deb && \
    dpkg -i fast-align_2016.05.31-1_amd64.deb && \
    rm fast-align_2016.05.31-1_amd64.deb

# Set the working directory
WORKDIR /app

# Clone eflomal repo and build the C binaries
RUN git clone https://github.com/robertostling/eflomal.git /app/eflomal && \
    cd /app/eflomal && \
    python -m pip install .

# Install required Python packages
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy the service code
COPY service.py .

# Expose the web server port
EXPOSE 8000

# Run the service
CMD ["uvicorn", "service:app", "--host", "0.0.0.0", "--port", "8000"]
