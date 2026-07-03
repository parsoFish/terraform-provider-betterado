# Fix Plan

> Checklist for WI-4. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_servicehook_storage_queue_pipelines is a framework resource under the mux WHEN TestAccServicehookStorageQueuePipelinesFramework_basic runs live (TF_ACC=1) THEN terraform apply creates the subscription; provider read-back asserts project_id, queue_name, account_key, and stage_state_changed_event attributes; ExpectNonEmptyPlan:false; destroy cleans up
- [x] AC2: GIVEN the live read-back inside the acceptance test WHEN CaptureLiveEvidence is called with label 'acceptance-resource' THEN .forge/live-evidence/acceptance-resource.json is written with a real ADO REST GET URL (https://dev.azure.com/.../_apis/hooks/subscriptions/<id>?api-version=7.1) and the subscription response body
