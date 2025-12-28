FROM python:3.14-slim

WORKDIR /app

RUN pip install uv

COPY pyproject.toml .python-version ./

RUN uv sync

COPY genall .

ENTRYPOINT ["uv", "run", "python", "-m", "genall"]
