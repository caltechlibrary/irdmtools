
> - Invenio RDM JSON API uses Bearer tokens for authentication, passed via HTTP headers or query parameters.
> - The API supports CRUD operations on drafts, records, files, and requests with versioning and access control.
> - Key endpoints include `/api/records`, `/api/drafts`, `/api/files`, and `/api/requests`, each supporting multiple HTTP methods.
> - Requests and responses follow RESTful conventions with JSON payloads, including metadata aligned with DataCite Schema v4.x.
> - API features pagination, sorting, filtering, and pretty-printing of JSON responses for exploration.

---

## Introduction

The Invenio RDM (Research Data Management) platform provides a robust JSON API for managing research data records, drafts, files, and user requests. This report synthesizes the official documentation into a structured, comprehensive Markdown table summarizing all API endpoints, their methods, descriptions, example requests, responses, and notes. The table is designed to be a ready-reference guide for developers and researchers interacting with Invenio RDM programmatically.

---

## Authentication and Authorization

The Invenio RDM API exclusively uses Bearer tokens for authentication. These tokens are generated in the user account settings and must be included in requests either via the `Authorization` HTTP header or as an `access_token` query parameter. The default scope is `user:email`, but custom scopes can be configured for granular access control. Authentication is mandatory for most endpoints, especially those involving drafts, records, and requests.

Example authentication via header:
```bash
curl -H "Authorization: Bearer API-TOKEN" https://inveniordm.example.com/api/records
```

Or via query parameter:
```bash
curl https://inveniordm.example.com/api/records?access_token=API-TOKEN
```

---

## API Endpoints Summary Table

| Endpoint Path                  | HTTP Method(s)       | Description                                                                                         | Example Request                                                                                   | Example Response (truncated)                                                                                  | Notes                                                                                             |
|-------------------------------|---------------------|-------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------|
| `/api/records`                | GET                 | Retrieve published records with optional pagination, sorting, and filtering                       | `GET /api/records?page=1&size=10&prettyprint=1`                                                  | `{ "hits": { "total": 100, "hits": [ { "id": "123", "metadata": { ... } } ] }`                                   | Supports `page`, `size`, `q` (query), `sort` parameters. Use `prettyprint=1` for formatted JSON. |
| `/api/records`                | POST                | Create a new record with metadata and files                                                       | `POST /api/records` with JSON body containing metadata and file info                              | `{ "id": "123", "metadata": { ... }, "files": { ... } }`                                         | Requires authentication. Files must be uploaded separately.                                      |
| `/api/records/{id}`           | GET                 | Retrieve a specific record by ID                                                                  | `GET /api/records/123`                                                                          | `{ "id": "123", "metadata": { ... }, "files": { ... } }`                                         | Returns detailed record metadata and associated files.                                           |
| `/api/records/{id}`           | PUT                 | Update an existing record                                                                         | `PUT /api/records/123` with JSON body containing updated metadata                                 | `{ "id": "123", "metadata": { ... } }`                                                          | Requires authentication. Only metadata can be updated; files require separate endpoints.         |
| `/api/records/{id}`           | DELETE              | Delete a record                                                                                   | `DELETE /api/records/123`                                                                       | `{ "status": "deleted" }`                                                                       | Requires authentication. Irreversible operation.                                               |
| `/api/records/{id}/versions`  | GET                 | List all versions of a record                                                                    | `GET /api/records/123/versions`                                                                | `{ "versions": [ { "id": "v1", "created": "2023-01-01T00:00:00Z" } ] }`                                       | Returns version history with timestamps.                                                         |
| `/api/records/{id}/versions`  | POST                | Create a new version of a record                                                                  | `POST /api/records/123/versions` with JSON body containing version metadata                        | `{ "id": "v2", "created": "2023-01-02T00:00:00Z" }`                                           | Requires authentication. Allows updating files in new versions.                                |
| `/api/drafts`                 | GET                 | List all drafts                                                                                   | `GET /api/drafts`                                                                              | `{ "hits": { "total": 10, "hits": [ { "id": "d123", "metadata": { ... } } ] }`                                   | Requires authentication. Supports pagination and filtering.                                     |
| `/api/drafts`                 | POST                | Create a new draft record                                                                         | `POST /api/drafts` with JSON body containing metadata and file info                              | `{ "id": "d123", "metadata": { ... } }`                                                        | Requires authentication. Files must be uploaded separately.                                      |
| `/api/drafts/{id}`            | GET                 | Retrieve a specific draft by ID                                                                    | `GET /api/drafts/d123`                                                                          | `{ "id": "d123", "metadata": { ... } }`                                                        | Returns detailed draft metadata.                                                                 |
| `/api/drafts/{id}`            | PUT                 | Update a draft record                                                                             | `PUT /api/drafts/d123` with JSON body containing updated metadata                                | `{ "id": "d123", "metadata": { ... } }`                                                        | Requires authentication. Only metadata can be updated; files require separate endpoints.         |
| `/api/drafts/{id}`            | DELETE              | Delete a draft record                                                                             | `DELETE /api/drafts/d123`                                                                       | `{ "status": "deleted" }`                                                                       | Requires authentication. Irreversible operation.                                               |
| `/api/drafts/{id}/publish`    | POST                | Publish a draft record                                                                             | `POST /api/drafts/d123/publish`                                                                | `{ "id": "123", "status": "published" }`                                                      | Requires authentication. Creates a published record from the draft.                             |
| `/api/drafts/{id}/files`      | GET                 | List files associated with a draft                                                               | `GET /api/drafts/d123/files`                                                                    | `{ "files": [ { "id": "f1", "filename": "file.txt" } ] }`                                     | Requires authentication. Returns file metadata.                                                 |
| `/api/drafts/{id}/files`      | POST                | Upload a file to a draft                                                                           | `POST /api/drafts/d123/files` with file data                                                   | `{ "id": "f1", "filename": "file.txt", "status": "uploaded" }`                                  | Requires authentication. Supports chunked uploads.                                            |
| `/api/drafts/{id}/files/{file_id}` | DELETE       | Delete a file from a draft                                                                         | `DELETE /api/drafts/d123/files/f1`                                                             | `{ "status": "deleted" }`                                                                       | Requires authentication. Irreversible operation.                                               |
| `/api/requests`                | GET                 | List all requests                                                                                  | `GET /api/requests`                                                                              | `{ "hits": { "total": 5, "hits": [ { "id": "r123", "status": "pending" } ] }`                                   | Requires authentication. Supports pagination and filtering.                                     |
| `/api/requests`                | POST                | Create a new request                                                                               | `POST /api/requests` with JSON body containing request details                                 | `{ "id": "r123", "status": "created" }`                                                       | Requires authentication. Requests can be accepted, declined, or canceled.                        |
| `/api/requests/{id}`           | GET                 | Retrieve a specific request by ID                                                                 | `GET /api/requests/r123`                                                                        | `{ "id": "r123", "status": "pending", "comments": [ ... ] }`                                    | Returns request details including comments and timeline.                                      |
| `/api/requests/{id}`           | PUT                 | Update a request (e.g., accept, decline, cancel)                                                 | `PUT /api/requests/r123` with JSON body containing updated status                               | `{ "id": "r123", "status": "accepted" }`                                                      | Requires authentication. Only certain users can update requests based on roles.                |
| `/api/requests/{id}/comments`  | GET                 | List comments on a request                                                                        | `GET /api/requests/r123/comments`                                                               | `{ "comments": [ { "id": "c1", "content": "comment text" } ] }`                              | Requires authentication. Returns all comments on the request.                                   |
| `/api/requests/{id}/comments`  | POST                | Add a comment to a request                                                                        | `POST /api/requests/r123/comments` with JSON body containing comment content                    | `{ "id": "c1", "content": "comment text" }`                                                    | Requires authentication. Comments can be updated and deleted.                                   |
| `/api/requests/{id}/comments/{comment_id}` | PUT, DELETE | Update or delete a comment on a request                                                          | `PUT /api/requests/r123/comments/c1` with updated content or `DELETE /api/requests/r123/comments/c1` | `{ "id": "c1", "content": "updated comment" }` or `{ "status": "deleted" }`                      | Requires authentication. Only comment owners or authorized users can modify.                    |
| `/api/records/{id}/access/links` | POST           | Create a secret link for a record with specific permissions                                      | `POST /api/records/123/access/links` with JSON body containing permission and expiration        | `{ "token": "secret-token", "expires_at": "2023-12-31T23:59:00Z" }`                         | Requires authentication. Used for sharing records with specific permissions and expiration.       |

---

## Key Notes and Additional Information

- **Authentication**: All endpoints except public record retrieval require a Bearer token. Tokens are generated in the user account settings and passed via HTTP headers or query parameters.
- **Pagination and Filtering**: Endpoints that return lists (e.g., `/api/records`, `/api/drafts`, `/api/requests`) support pagination with `page` and `size` parameters, and filtering via `q` (query) parameter.
- **Sorting**: Records can be sorted using the `sort` parameter with options like "bestmatch," "newest," "mostviewed," and "mostdownloaded."
- **Pretty Printing**: Adding `prettyprint=1` to the query string formats JSON responses for easier readability in browsers.
- **Versioning**: Records support versioning, allowing creation of new versions with updated files and metadata.
- **File Management**: Files are managed separately from records and drafts, with endpoints supporting upload, deletion, and listing.
- **Request Management**: Requests can be created, updated, and commented on, with detailed timelines and status tracking.
- **Secret Links**: The API supports creating secret links with specific permissions and expiration dates for secure sharing of records.

---

## Conclusion

The Invenio RDM JSON API offers a comprehensive, well-documented interface for managing research data records, drafts, files, and user requests. The API follows RESTful conventions with clear endpoints for CRUD operations, versioning, and access control. Authentication via Bearer tokens ensures secure access, and the documentation provides practical examples for interacting with each endpoint. The API's support for pagination, sorting, filtering, and pretty-printing enhances usability for developers and researchers. This structured summary table serves as a ready-reference guide for integrating and utilizing the Invenio RDM API effectively.
