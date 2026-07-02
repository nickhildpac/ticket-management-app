from __future__ import annotations

import importlib
import json
import sys
from pathlib import Path
from typing import Any

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

POSTMAN_SCHEMA = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
OUTPUT_PATH = Path(__file__).resolve().parents[1] / "postman_collection.json"
app = importlib.import_module("app.main").app
PUBLIC_OPERATION_IDS = {
    "health_health_get",
    "api_health_api_v1_health_get",
    "register_user_api_v1_users_post",
    "login_api_v1_auth_login_post",
    "refresh_token_api_v1_auth_refresh_get",
    "logout_api_v1_auth_logout_get",
}


def _schema_example(schema: dict[str, Any], components: dict[str, Any]) -> Any:
    if "$ref" in schema:
        name = schema["$ref"].rsplit("/", maxsplit=1)[-1]
        return _schema_example(components.get(name, {}), components)

    if "anyOf" in schema:
        choices = [item for item in schema["anyOf"] if item.get("type") != "null"]
        return _schema_example(choices[0], components) if choices else None

    if "allOf" in schema and schema["allOf"]:
        return _schema_example(schema["allOf"][0], components)

    schema_type = schema.get("type")
    if schema_type == "object" or "properties" in schema:
        return {
            name: _schema_example(prop_schema, components)
            for name, prop_schema in schema.get("properties", {}).items()
        }
    if schema_type == "array":
        return [_schema_example(schema.get("items", {}), components)]
    if schema.get("format") == "email":
        return "user@example.com"
    if schema.get("format") == "uuid":
        return "00000000-0000-0000-0000-000000000000"
    if schema_type == "integer":
        return 0
    if schema_type == "number":
        return 0
    if schema_type == "boolean":
        return False
    if "enum" in schema and schema["enum"]:
        return schema["enum"][0]
    return "string"


def _postman_url(path: str, parameters: list[dict[str, Any]]) -> dict[str, Any]:
    variable_names = []
    raw_path = path
    for parameter in parameters:
        if parameter.get("in") == "path":
            name = parameter["name"]
            variable_names.append({"key": name, "value": f"<{name}>"})
            raw_path = raw_path.replace(f"{{{name}}}", f":{name}")

    query = [
        {
            "key": parameter["name"],
            "value": "",
            "disabled": not parameter.get("required", False),
        }
        for parameter in parameters
        if parameter.get("in") == "query"
    ]

    url: dict[str, Any] = {
        "raw": "{{baseUrl}}" + raw_path,
        "host": ["{{baseUrl}}"],
        "path": [part for part in raw_path.lstrip("/").split("/") if part],
    }
    if query:
        url["query"] = query
    if variable_names:
        url["variable"] = variable_names
    return url


def _request_body(operation: dict[str, Any], components: dict[str, Any]) -> dict[str, Any] | None:
    content = operation.get("requestBody", {}).get("content", {})
    json_body = content.get("application/json")
    if not json_body:
        return None

    example = _schema_example(json_body.get("schema", {}), components)
    return {
        "mode": "raw",
        "raw": json.dumps(example, indent=2),
        "options": {"raw": {"language": "json"}},
    }


def _auth_for(operation: dict[str, Any]) -> dict[str, Any] | None:
    if operation.get("operationId") in PUBLIC_OPERATION_IDS:
        return None
    return {
        "type": "bearer",
        "bearer": [{"key": "token", "value": "{{accessToken}}", "type": "string"}],
    }


def _operation_item(
    method: str,
    path: str,
    operation: dict[str, Any],
    components: dict[str, Any],
) -> dict[str, Any]:
    parameters = operation.get("parameters", [])
    request: dict[str, Any] = {
        "method": method.upper(),
        "header": [{"key": "Content-Type", "value": "application/json"}],
        "url": _postman_url(path, parameters),
        "description": operation.get("description") or operation.get("summary", ""),
    }

    body = _request_body(operation, components)
    if body:
        request["body"] = body

    auth = _auth_for(operation)
    if auth:
        request["auth"] = auth

    item: dict[str, Any] = {
        "name": operation.get("summary") or operation.get("operationId") or f"{method} {path}",
        "request": request,
        "response": [],
    }

    if operation.get("operationId") == "login_api_v1_auth_login_post":
        item["event"] = [
            {
                "listen": "test",
                "script": {
                    "type": "text/javascript",
                    "exec": [
                        "const body = pm.response.json();",
                        "if (body.access_token) {",
                        "  pm.collectionVariables.set('accessToken', body.access_token);",
                        "}",
                    ],
                },
            }
        ]

    return item


def main() -> None:
    openapi = app.openapi()
    components = openapi.get("components", {}).get("schemas", {})
    folders: dict[str, list[dict[str, Any]]] = {}

    for path, path_item in sorted(openapi["paths"].items()):
        for method, operation in sorted(path_item.items()):
            if method.lower() not in {"get", "post", "put", "patch", "delete"}:
                continue
            tag = operation.get("tags", ["General"])[0]
            folders.setdefault(tag, []).append(
                _operation_item(method, path, operation, components)
            )

    collection = {
        "info": {
            "name": openapi.get("info", {}).get("title", "Ticket Management API"),
            "description": "Generated from the FastAPI OpenAPI schema.",
            "schema": POSTMAN_SCHEMA,
        },
        "item": [{"name": name, "item": items} for name, items in sorted(folders.items())],
        "variable": [
            {"key": "baseUrl", "value": "http://localhost:8081"},
            {"key": "accessToken", "value": ""},
        ],
    }

    OUTPUT_PATH.write_text(json.dumps(collection, indent=2) + "\n", encoding="utf-8")
    print(f"Wrote {OUTPUT_PATH}")


if __name__ == "__main__":
    main()
