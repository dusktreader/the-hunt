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

r = httpx.get("http://localhost:4000/v1/companies")
r.raise_for_status()

companies = r.json()["companies"]
for company in companies:
    logger.info(f"Found: {company['name']}")

IPython.embed()
