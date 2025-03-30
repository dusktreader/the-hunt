#!/usr/bin/env -S uv run --script
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

users = [
    dict(name="The Dude", email="the.dude@abides.com", password="thedudeabides"),
    dict(name="Walter Sobchak", email="walter@sobchak-security.com", password="vietnamvet"),
    dict(name="Donnie", email="donniesurfs@yahoo.com", password="iamthewalrus"),
    dict(name="Maude", email="mauddie@avant-guard.com", password="goodmanandthorough"),
]

r = httpx.get("http://localhost:4000/v1/users")
r.raise_for_status()

for user in r.json()["users"]:
    logger.info(f"Deleting: {user['name']}")
    r = httpx.delete(f"http://localhost:4000/v1/users/{user['id']}")
    r.raise_for_status()

for user in users:
    r = httpx.post("http://localhost:4000/v1/users", json=user)
    r.raise_for_status()
    logger.info(f"Created: {r.json()}")

IPython.embed()
