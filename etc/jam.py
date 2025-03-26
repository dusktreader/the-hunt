#!/usr/bin/env -S uv run --script
# /// script
# dependencies = [
#   "ipython",
#   "httpx",
#   "loguru",
# ]
# ///

from typing import Any

import httpx
import IPython
from loguru import logger

users = [
    dict(name="The Dude", email="the.dude@abides.com", password="thedudeabides"),
    dict(name="Walter Sobchak", email="walter@sobchak-security.com", password="vietnamvet"),
    dict(name="Donnie", email="donniesurfs@yahoo.com", password="iamthewalrus"),
    dict(name="Maude", email="mauddie@avant-guard.com", password="goodmanandthorough"),
]

companies = [
    dict(name="Close", url="https://close.com", tech_stack=["Python", "PostgreSQL", "Kubernetes"]),
    dict(name="Clever", url="https://clever.com", tech_stack=["Go", "Kubernetes"]),
    dict(name="iSpotTV", url="https://ispot.com", tech_stack=["Java", "Mysql", "Kubernetes"]),
    dict(name="Canonical", url="https://canonical.org", tech_stack=["Python", "Go", "Kubernetes"]),
]

def delete_all_companies(client: httpx.Client):
    logger.info("Deleting all companies")
    r = client.get("/companies")
    r.raise_for_status()

    for company in r.json()["companies"]:
        logger.info(f"Deleting: {company['name']}")
        r = client.delete(f"/companies/{company['id']}")
        r.raise_for_status()


def insert_companies(client: httpx.Client):
    logger.info("Inserting companies")
    for company in companies:
        r = client.post("/companies", json=company)
        r.raise_for_status()
        logger.info(f"Created: {r.json()}")


def delete_all_users(client: httpx.Client):
    logger.info("Deleting all users")
    r = client.get("/users")
    r.raise_for_status()

    for user in r.json()["users"]:
        logger.info(f"Deleting: {user['name']}")
        r = client.delete(f"/users/{user['id']}")
        r.raise_for_status()


def insert_users(client: httpx.Client):
    logger.info("Inserting users")
    for user in users:
        r = client.post("/users", json=user)
        r.raise_for_status()
        logger.info(f"Created: {r.json()}")


def auth_me(token: str) -> dict[str, dict[str, str]]:
    return dict(headers=dict(Authorization=f"Bearer {token}"))


def main():
    logger.info("Getting admin auth token")
    r = httpx.post("http://localhost:4000/v1/login", json=dict(email="admin@the-hunt.dev", password="admin"))
    r.raise_for_status()
    admin_token = r.json()["auth"]["token"]
    logger.info(f"Got admin token: {admin_token}")

    with httpx.Client(
        base_url="http://localhost:4000/v1",
        headers=dict(Authorization=f"Bearer {admin_token}")
    ) as admin_client:
        delete_all_companies(admin_client)
        insert_companies(admin_client)
        delete_all_users(admin_client)
        insert_users(admin_client)

        with httpx.Client(
            base_url="http://localhost:4000/v1",
        ) as client:
            IPython.embed()


if __name__ == "__main__":
    main()
