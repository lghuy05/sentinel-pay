FROM python:3.12-slim

ENV PYTHONUNBUFFERED=1

WORKDIR /app

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY inference ./inference
COPY model_artifacts ./model_artifacts

WORKDIR /app/inference

EXPOSE 8091

CMD ["python", "main.py"]
