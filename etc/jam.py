# /// script
# dependencies = [
#   "ipython",
#   "httpx",
#   "loguru",
# ]
# ///

import httpx
import IPython
from loguru import logger

companies = [
    dict(name="Close", url="https://close.com", tech_stack=["Python", "PostgreSQL", "Kubernetes"]),
    dict(name="Clever", url="https://clever.com", tech_stack=["Go", "Kubernetes"]),
    dict(name="iSpotTV", url="https://ispot.com", tech_stack=["Java", "Mysql", "Kubernetes"]),
    dict(name="Canonical", url="https://canonical.org", tech_stack=["Python", "Go", "Kubernetes"]),
]

r = httpx.get("http://localhost:4000/v1/companies")
r.raise_for_status()

for company in r.json()["companies"]:
    logger.info(f"Deleting: {company['name']}")
    r = httpx.delete(f"http://localhost:4000/v1/companies/{company['id']}")
    r.raise_for_status()

for company in companies:
    r = httpx.post(
        "http://localhost:4000/v1/companies",
        json=company,
    )
    r.raise_for_status()
    logger.info(f"Created: {r.json()}")

IPython.embed()
