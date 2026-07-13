### Task 14: Final Polish — Full Stack Integration Test

- [ ] **Step 1: Run Go backend tests**
```bash
cd backend && go test ./... -count=1
```

- [ ] **Step 2: Run frontend build**
```bash
cd frontend && npm run build
```

- [ ] **Step 3: Try Docker build (if docker is available)**
```bash
docker build -t hermes-filebrowser:latest .
```

If Docker is not available, skip this step.

- [ ] **Step 4: Final commit**
```bash
git add -A
git commit -m "chore: final integration fixes"
```
