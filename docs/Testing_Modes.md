# Testing Modes

## API/Integration
```bash
cd backend && go test ./tests/API_tests/... -count=1
```

## Unit Tests
```bash
cd backend && go test ./tests/unit_tests/... -count=1
```

## Full Stack (Compose)
```bash
./run_tests.sh
```

## Focused Adversarial Checks
- Evidence spoofing with cross-booking IDs must fail.
- Consultation version collisions across bookings must not occur.
- MIME bypass uploads (for example text payloads) must return `415`.
