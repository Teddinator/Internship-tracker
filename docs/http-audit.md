# HTTP Audit

## GET /applications/

- Method: GET
- Safe: Yes
- Idempotent: Yes
- Request body: None
- Success: 200 OK
- Response: application/json
- Notes: Returns an empty array if no application exist.

## POST /applications/

- Method: POST
- Safe: No
- Idempotent: No
- Request body: JSON
- Success: 201 Created
- Headers:
    - Content-Type: application/json
    - Location: /applications/{id}
- Errors:
    - 400 invalid input
    - 404 company does not exist

## DELETE /applications/{id}

- Method: DELETE
- Safe: No
- Idempotent: Yes
- Success: 204 No Content
- Errors:
    - 400 invalid id
    - 404 application not found