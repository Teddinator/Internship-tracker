# HTTP Audit

## GET /applications/

- Method: GET
- Safe: Yes
- Idempotent: Yes
- Request body: None
- Success: 200 OK
- Response: application/json
- Headers:
    - Content-Type: application/json
- Notes: Returns an empty array if no application exist.

## POST /applications/

- Method: POST
- Safe: No
- Idempotent: No
- Request body: JSON
- Success: 201 Created
- Response: application/json
- Headers:
    - Content-Type: application/json
    - Location: /applications/{id}
- Errors:
    - 400 Bad Request: invalid request body or field values
    - 404 Not Found: company does not exist

## GET /applications/{id}

- Method: GET
- Safe: Yes
- Idempotent: Yes
- Request body: None
- Success: 200 OK
- Response: application/json
- Headers:
    - Content-Type: application/json
- Errors:
    - 400 Bad Request: invalid application id
    - 404 Not Found: application not found

## PUT /applications/{id}

- Method: PUT
- Safe: No
- Idempotent: Yes
- Request body: JSON
- Success: 200 OK
- Response: application/json
- Headers:
    - Content-Type: application/json
- Errors:
    - 400 Bad Request: invalid request body or field values
    - 404 Not Found: application not found

## DELETE /applications/{id}

- Method: DELETE
- Safe: No
- Idempotent: Yes
- Request body: None
- Success: 204 No Content
- Errors:
    - 400 Bad Request: invalid application id
    - 404 Not Found: application not found

## GET /applications/export.csv

- Method: GET
- Safe: Yes
- Idempotent: Yes
- Request body: None
- Success: 200 OK
- Response: text/csv
- Headers:
    - Content-Type: text/csv; charset=utf-8
    - Content-Disposition: attachment; filename="applications.csv"

## GET /companies/

- Method: GET
- Safe: Yes
- Idempotent: Yes
- Request body: None
- Success: 200 OK
- Response: application/json
- Headers:
    - Content-Type: application/json

## POST /companies/

- Method: POST
- Safe: No
- Idempotent: No
- Request body: JSON
- Success: 201 Created
- Response: application/json
- Headers:
    - Content-Type: application/json
    - Location: /companies/{id}
- Errors:
    - 400 Bad Request: invalid request body or field values

## GET /companies/{id}

- Method: GET
- Safe: Yes
- Idempotent: Yes
- Request body: None
- Success: 200 OK
- Response: application/json
- Headers:
    - Content-Type: application/json
- Errors:
    - 400 Bad Request: invalid company id
    - 404 Not Found: company not found

## PUT /companies/{id}

- Method: PUT
- Safe: No
- Idempotent: Yes
- Request body: JSON
- Success: 200 OK
- Response: application/json
- Headers:
    - Content-Type: application/json
- Errors:
    - 400 Bad Request: invalid request body or field values
    - 404 Not Found: company not found

## DELETE /companies/{id}

- Method: DELETE
- Safe: No
- Idempotent: Yes
- Request body: None
- Success: 204 No Content
- Errors:
    - 400 Bad Request: invalid company id
    - 404 Not Found: company not found
    - 409 Conflict: company cannot be deleted while applications reference it