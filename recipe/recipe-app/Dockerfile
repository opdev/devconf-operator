# Use an official Python runtime as a parent image
FROM --platform=linux/amd64 python:3.7

# Set the working directory
WORKDIR /app

# Create the /database directory
RUN mkdir -p /database

# Copy the current directory contents into the container at /app
COPY . /app

# Install any needed packages specified in requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

# Run the application with Gunicorn
CMD ["gunicorn", "-w", "4", "-b", "0.0.0.0:5000", "app:app"]