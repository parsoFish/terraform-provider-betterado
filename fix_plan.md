# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a valid project_id and path are provided WHEN dataReleaseFolderRead is called and GetFolders returns a matching folder THEN the resource data has Id set to path, description populated, and no error is returned
- [x] AC2: GIVEN a valid project_id and path are provided WHEN dataReleaseFolderRead is called and GetFolders returns an empty slice THEN an error is returned containing the path so the caller knows the folder was not found
