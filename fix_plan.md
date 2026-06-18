# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN no gap matrix document exists at docs/release-definition-permissions-gap-matrix.md WHEN WI-1 completes THEN docs/release-definition-permissions-gap-matrix.md exists and lists every ReleaseManagement2 permission bit (name + bit value + writable flag)
- [x] AC2: GIVEN the gap matrix is written WHEN the writable column is reviewed THEN every permission bit that ADO allows to be set via SetAccessControlEntries is marked Writable=yes
