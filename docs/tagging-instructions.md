# Tagging Instructions for v1.1.0

## Contract Freeze Tag

To create the v1.1.0 tag for the contract freeze:

```bash
# Create annotated tag
git tag -a v1.1.0 -m "Contract Freeze v1.1.0 - Stage 1

- All endpoints return schema_version: ds.v1
- Self-status includes nowMs timestamp
- Discovery endpoints implemented
- OpenAPI documentation complete
- CI validation gates in place

Contract frozen as of 2025-09-29"

# Push tag to remote (when ready)
git push origin v1.1.0
```

## Tag Contents

The v1.1.0 tag represents:
- ✅ Schema version `ds.v1` on all endpoints
- ✅ Wrapped responses for array endpoints
- ✅ Discovery endpoints operational
- ✅ Complete OpenAPI documentation
- ✅ CI/CD validation workflows
- ✅ Contract version documentation

## Verification Before Tagging

Run these checks before creating the tag:

```bash
# 1. Run tests
go test ./...

# 2. Build verification
go build ./cmd/ds

# 3. Run Stage 0 verification
./scripts/verify-stage0.sh

# 4. Check contract documentation
cat docs/contracts/VERSION.md

# 5. Verify CI workflow
cat .github/workflows/validate-contracts.yml
```

## After Tagging

1. Update release notes
2. Notify other agents of contract freeze
3. No breaking changes allowed after this point