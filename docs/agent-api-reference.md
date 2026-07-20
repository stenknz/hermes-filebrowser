# FileBrowser API Reference for Hermes Agent

**Server:** `http://filebrowser:8282`  
**Auth:** `Authorization: Bearer fb_<api-token>`

---

## File Operations

### List Directory
```
GET /api/files?path=<path>
GET /api/files?path=<path>/     # trailing slash also works
```
→ `{"data": [{"name":"...", "path":"...", "size":N, "isDir":bool, "modTime":"..."}]}`

### Create Directory
```
POST /api/files/dir?path=<path>
POST /api/files/mkdir {"path": "<path>"}
POST /api/files/file {"path": "<path>/", "content": ""}
POST /api/files/file {"path": "<path>", "type": "dir"}
```
All create intermediate parents automatically.

### Write File — Text
```
POST /api/files/file {"path": "<path>", "content": "text here"}
```

### Write File — Binary (base64)
```
POST /api/files/file {"path": "<path>", "content": "<base64>", "encoding": "base64"}
POST /api/files/file {"path": "<path>", "content": "<base64>", "base64": true}
```

### Write File — Binary (raw bytes, no encoding)
```
POST /api/files/write?path=<path>
Content-Type: application/octet-stream
<body bytes>
```
Bytes are written as-is — no UTF-8 re-encoding, no corruption.

### Upload File (multipart)
```
POST /api/upload?path=<dir>
Content-Type: multipart/form-data
(file in "file" field)
```

### Read File
```
GET /api/files/raw?path=<path>
```

### Get File Metadata (stat)
```
GET /api/files/stat?path=<path>
```
→ `{"name":"...","path":"...","size":N,"isDir":bool,"modTime":"...","mode":"..."}`

### Rename / Move
```
PATCH /api/files/file {"source": "<old>", "destination": "<new>"}
POST /api/files/rename {"oldPath": "<old>", "newPath": "<new>"}
PUT  /api/files/rename {"oldPath": "<old>", "newPath": "<new>"}
POST /api/files/rename {"source": "<old>", "destination": "<new>"}
POST /api/files/rename {"from": "<old>", "to": "<new>"}
POST /api/files/move {"source": "<old>", "destination": "<new>"}
```
All accept any field name pair: `oldPath`/`newPath`, `source`/`destination`, `from`/`to`. Creates intermediate parent dirs automatically.

### Copy
```
POST /api/files/copy {"source": "<src>", "destination": "<dst>"}
```

### Delete
```
DELETE /api/files?path=<path>
DELETE /api/files/file?path=<path>
DELETE /api/files/dir?path=<path>
POST   /api/files/delete {"path": "<path>"}
POST   /api/files/delete {"names": ["<path1>", "<path2>"]}
```

---

## Search

```
GET /api/search?query=<text>&path=<base-dir>
GET /api/search?q=<text>&path=<base-dir>
```
Searches filenames only (not content) within the given directory.

---

## User & Token Management (admin only)

```
GET  /api/me                          → current user info
GET  /api/users                       → list users
POST /api/users                       → create user
POST /api/users/delete {"id": N}      → delete user
GET  /api/tokens                      → list my API tokens
POST /api/tokens {"name": "..."}      → create API token
POST /api/tokens/delete {"id": N}     → revoke API token
```

---

## Key Behavior Notes

- **Base64 encoding:** When `"encoding":"base64"` or `"base64":true` is set, the server decodes base64 before writing. Without it, content is written as raw UTF-8 string bytes.
- **Auto-create parents:** All write operations (`POST /api/files/file`, `POST /api/files/mkdir`, `POST /api/files/write`, etc.) automatically create intermediate directories if they don't exist.
- **Path traversal:** Any path containing `..` is rejected. All paths are resolved relative to `FB_ROOT`.
- **Hidden files:** Dotfiles and `filebrowser.db*` files are hidden from listing and protected from modification.
- **API tokens bypass CSRF:** Session tokens (UUIDs) do NOT bypass CSRF — only API tokens (`fb_` prefix) can write without CSRF headers.
