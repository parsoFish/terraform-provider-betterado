# Member Entitlement Management — Gap Matrix

> **Initiative:** INIT-2026-07-01-migrate-framework-member-entitlement
> **Source:** ADO Member Entitlement Management REST API v7.1
> **Compared against:** SDKv2 schemas for `betterado_user_entitlement`, `betterado_group_entitlement`, `betterado_service_principal_entitlement`
> **Gap statuses:** `covered` | `gap-deferred` | `covered-this-cycle`
> **Per initiative scope:** new API features are out of scope — undiscovered gaps are `gap-deferred`.

---

## 1. `betterado_user_entitlement` — `UserEntitlement` struct

The `UserEntitlement` struct embeds `EntitlementBase` fields and adds a `User` (`GraphUser`) reference and a deprecated `Extensions` field.

### 1a. Top-level `UserEntitlement` fields

| ADO REST field (`json` tag) | SDK Go struct field | Terraform attribute | Writable? | Gap status |
|---|---|---|---|---|
| `accessLevel` | `AccessLevel *licensing.AccessLevel` | _(nested — see §1b)_ | yes | covered |
| `dateCreated` | `DateCreated *azuredevops.Time` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `groupAssignments` | `GroupAssignments *[]GroupEntitlement` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `id` | `Id *uuid.UUID` | resource ID (implicit) | yes (create only) | covered |
| `lastAccessedDate` | `LastAccessedDate *azuredevops.Time` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `projectEntitlements` | `ProjectEntitlements *[]ProjectEntitlement` | _(not exposed)_ | yes | gap-deferred | — re-evaluation: `complexity-then`
| `extensions` | `Extensions *[]Extension` | _(not exposed; deprecated)_ | yes | gap-deferred | — re-evaluation: `non-declarative-forever`
| `user` | `User *graph.GraphUser` | _(nested — see §1c)_ | gap-open | covered |

### 1b. `accessLevel` (`licensing.AccessLevel`) sub-fields

| ADO REST field | SDK field | Terraform attribute | Writable? | Gap status |
|---|---|---|---|---|
| `accessLevel.accountLicenseType` | `AccountLicenseType *licensing.AccountLicenseType` | `account_license_type` | yes | covered |
| `accessLevel.assignmentSource` | `AssignmentSource *licensing.AssignmentSource` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `accessLevel.licenseDisplayName` | `LicenseDisplayName *string` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `accessLevel.licensingSource` | `LicensingSource *licensing.LicensingSource` | `licensing_source` | yes | covered |
| `accessLevel.msdnLicenseType` | `MsdnLicenseType *licensing.MsdnLicenseType` | _(not exposed)_ | yes | gap-deferred | — re-evaluation: `complexity-then`
| `accessLevel.status` | `Status *accounts.AccountUserStatus` | _(not exposed; used internally for deleted check)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `accessLevel.statusMessage` | `StatusMessage *string` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`

### 1c. `user` (`graph.GraphUser`) sub-fields

| ADO REST field | SDK field | Terraform attribute | Writable? | Gap status |
|---|---|---|---|---|
| `user.descriptor` | `Descriptor *string` | `descriptor` (computed) | no (readonly) | covered |
| `user.displayName` | `DisplayName *string` | _(not exposed)_ | no | gap-deferred | — re-evaluation: `non-declarative-forever`
| `user.url` | `Url *string` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `user.legacyDescriptor` | `LegacyDescriptor *string` | _(not exposed)_ | no (internal) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `user.origin` | `Origin *string` | `origin` (computed+optional) | yes (create only) | covered |
| `user.originId` | `OriginId *string` | `origin_id` (computed+optional) | yes (create only) | covered |
| `user.subjectKind` | `SubjectKind *string` | _(not exposed; hardcoded `"user"`)_ | no | gap-deferred | — re-evaluation: `non-declarative-forever`
| `user.domain` | `Domain *string` | _(not exposed)_ | no | gap-deferred | — re-evaluation: `non-declarative-forever`
| `user.mailAddress` | `MailAddress *string` | _(not exposed)_ | no | gap-deferred | — re-evaluation: `non-declarative-forever`
| `user.principalName` | `PrincipalName *string` | `principal_name` (computed+optional) | yes (create only) | covered |
| `user.directoryAlias` | `DirectoryAlias *string` | _(not exposed)_ | no | gap-deferred | — re-evaluation: `non-declarative-forever`
| `user.isDeletedInOrigin` | `IsDeletedInOrigin *bool` | _(not exposed; used internally)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `user.metaType` | `MetaType *string` | _(not exposed)_ | no | gap-deferred | — re-evaluation: `non-declarative-forever`
| `user._links` | `Links interface{}` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`

---

## 2. `betterado_group_entitlement` — `GroupEntitlement` struct

The `GroupEntitlement` struct holds a group identity (`GraphGroup`), a license rule (`AccessLevel`), and project membership.

### 2a. Top-level `GroupEntitlement` fields

| ADO REST field (`json` tag) | SDK Go struct field | Terraform attribute | Writable? | Gap status |
|---|---|---|---|---|
| `extensionRules` | `ExtensionRules *[]Extension` | _(not exposed; deprecated)_ | yes | gap-deferred | — re-evaluation: `non-declarative-forever`
| `group` | `Group *graph.GraphGroup` | _(nested — see §2b)_ | gap-open | covered |
| `id` | `Id *uuid.UUID` | resource ID (implicit) | yes (create only) | covered |
| `lastExecuted` | `LastExecuted *azuredevops.Time` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `licenseRule` | `LicenseRule *licensing.AccessLevel` | _(nested — see §2c)_ | yes | covered |
| `members` | `Members *[]UserEntitlement` | _(not exposed; create-only hint in SDK)_ | yes (create only) | gap-deferred | — re-evaluation: `complexity-then`
| `projectEntitlements` | `ProjectEntitlements *[]ProjectEntitlement` | _(not exposed)_ | yes | gap-deferred | — re-evaluation: `complexity-then`
| `status` | `Status *licensingrule.GroupLicensingRuleStatus` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`

### 2b. `group` (`graph.GraphGroup`) sub-fields

| ADO REST field | SDK field | Terraform attribute | Writable? | Gap status |
|---|---|---|---|---|
| `group.descriptor` | `Descriptor *string` | `descriptor` (computed) | no (readonly) | covered |
| `group.displayName` | `DisplayName *string` | `display_name` (computed+optional) | yes (create only) | covered |
| `group.url` | `Url *string` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `group.legacyDescriptor` | `LegacyDescriptor *string` | _(not exposed)_ | no (internal) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `group.origin` | `Origin *string` | `origin` (computed+optional) | yes (create only) | covered |
| `group.originId` | `OriginId *string` | `origin_id` (computed+optional) | yes (create only) | covered |
| `group.subjectKind` | `SubjectKind *string` | _(not exposed; hardcoded `"group"`)_ | no | gap-deferred | — re-evaluation: `non-declarative-forever`
| `group.domain` | `Domain *string` | _(not exposed)_ | no | gap-deferred | — re-evaluation: `non-declarative-forever`
| `group.mailAddress` | `MailAddress *string` | _(not exposed)_ | no | gap-deferred | — re-evaluation: `non-declarative-forever`
| `group.principalName` | `PrincipalName *string` | `principal_name` (computed) | no (readonly) | covered |
| `group.description` | `Description *string` | _(not exposed)_ | yes | gap-deferred | — re-evaluation: `complexity-then`
| `group.isDeleted` | `IsDeleted *bool` | _(not exposed; used internally for deleted check)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `group._links` | `Links interface{}` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`

### 2c. `licenseRule` (`licensing.AccessLevel`) sub-fields

| ADO REST field | SDK field | Terraform attribute | Writable? | Gap status |
|---|---|---|---|---|
| `licenseRule.accountLicenseType` | `AccountLicenseType *licensing.AccountLicenseType` | `account_license_type` | yes | covered |
| `licenseRule.assignmentSource` | `AssignmentSource *licensing.AssignmentSource` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `licenseRule.licenseDisplayName` | `LicenseDisplayName *string` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `licenseRule.licensingSource` | `LicensingSource *licensing.LicensingSource` | `licensing_source` | yes | covered |
| `licenseRule.msdnLicenseType` | `MsdnLicenseType *licensing.MsdnLicenseType` | _(not exposed)_ | yes | gap-deferred | — re-evaluation: `complexity-then`
| `licenseRule.status` | `Status *accounts.AccountUserStatus` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `licenseRule.statusMessage` | `StatusMessage *string` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`

---

## 3. `betterado_service_principal_entitlement` — `ServicePrincipalEntitlement` struct

The `ServicePrincipalEntitlement` struct embeds `EntitlementBase` fields and adds a `ServicePrincipal` (`GraphServicePrincipal`) reference.

### 3a. Top-level `ServicePrincipalEntitlement` fields

| ADO REST field (`json` tag) | SDK Go struct field | Terraform attribute | Writable? | Gap status |
|---|---|---|---|---|
| `accessLevel` | `AccessLevel *licensing.AccessLevel` | _(nested — see §3b)_ | yes | covered |
| `dateCreated` | `DateCreated *azuredevops.Time` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `groupAssignments` | `GroupAssignments *[]GroupEntitlement` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `id` | `Id *uuid.UUID` | resource ID (implicit) | yes (create only) | covered |
| `lastAccessedDate` | `LastAccessedDate *azuredevops.Time` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `projectEntitlements` | `ProjectEntitlements *[]ProjectEntitlement` | _(not exposed)_ | yes | gap-deferred | — re-evaluation: `complexity-then`
| `servicePrincipal` | `ServicePrincipal *graph.GraphServicePrincipal` | _(nested — see §3c)_ | gap-open | covered |

### 3b. `accessLevel` (`licensing.AccessLevel`) sub-fields

| ADO REST field | SDK field | Terraform attribute | Writable? | Gap status |
|---|---|---|---|---|
| `accessLevel.accountLicenseType` | `AccountLicenseType *licensing.AccountLicenseType` | `account_license_type` | yes | covered |
| `accessLevel.assignmentSource` | `AssignmentSource *licensing.AssignmentSource` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `accessLevel.licenseDisplayName` | `LicenseDisplayName *string` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `accessLevel.licensingSource` | `LicensingSource *licensing.LicensingSource` | `licensing_source` | yes | covered |
| `accessLevel.msdnLicenseType` | `MsdnLicenseType *licensing.MsdnLicenseType` | _(not exposed)_ | yes | gap-deferred | — re-evaluation: `complexity-then`
| `accessLevel.status` | `Status *accounts.AccountUserStatus` | _(not exposed; used internally for deleted check)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `accessLevel.statusMessage` | `StatusMessage *string` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`

### 3c. `servicePrincipal` (`graph.GraphServicePrincipal`) sub-fields

| ADO REST field | SDK field | Terraform attribute | Writable? | Gap status |
|---|---|---|---|---|
| `servicePrincipal.descriptor` | `Descriptor *string` | `descriptor` (computed) | no (readonly) | covered |
| `servicePrincipal.displayName` | `DisplayName *string` | `display_name` (computed) | no (readonly) | covered |
| `servicePrincipal.url` | `Url *string` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `servicePrincipal.legacyDescriptor` | `LegacyDescriptor *string` | _(not exposed)_ | no (internal) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `servicePrincipal.origin` | `Origin *string` | `origin` (computed+optional) | yes (create only) | covered |
| `servicePrincipal.originId` | `OriginId *string` | `origin_id` (required) | yes (create only) | covered |
| `servicePrincipal.subjectKind` | `SubjectKind *string` | _(not exposed; hardcoded `"servicePrincipal"`)_ | no | gap-deferred | — re-evaluation: `non-declarative-forever`
| `servicePrincipal.domain` | `Domain *string` | _(not exposed)_ | no | gap-deferred | — re-evaluation: `non-declarative-forever`
| `servicePrincipal.mailAddress` | `MailAddress *string` | _(not exposed)_ | no | gap-deferred | — re-evaluation: `non-declarative-forever`
| `servicePrincipal.principalName` | `PrincipalName *string` | _(not exposed)_ | no | gap-deferred | — re-evaluation: `non-declarative-forever`
| `servicePrincipal.directoryAlias` | `DirectoryAlias *string` | _(not exposed)_ | no | gap-deferred | — re-evaluation: `non-declarative-forever`
| `servicePrincipal.isDeletedInOrigin` | `IsDeletedInOrigin *bool` | _(not exposed; used internally for deleted check)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`
| `servicePrincipal.metaType` | `MetaType *string` | _(not exposed)_ | no | gap-deferred | — re-evaluation: `non-declarative-forever`
| `servicePrincipal.applicationId` | `ApplicationId *string` | _(not exposed)_ | no | gap-deferred | — re-evaluation: `non-declarative-forever`
| `servicePrincipal._links` | `Links interface{}` | _(not exposed)_ | no (readonly) | gap-deferred | — re-evaluation: `non-declarative-forever`

---

## Summary of writable gaps

All writable gaps not covered in the current provider schema are marked `gap-deferred` per the initiative's "Not in scope" clause (no new API features in this cycle). The following writable fields are deferred for future cycles:

| Resource | Deferred writable field(s) |
|---|---|
| `user_entitlement` | `projectEntitlements`, `extensions` (deprecated), `accessLevel.msdnLicenseType` |
| `group_entitlement` | `members`, `projectEntitlements`, `extensionRules` (deprecated), `licenseRule.msdnLicenseType`, `group.description` |
| `service_principal_entitlement` | `projectEntitlements`, `accessLevel.msdnLicenseType` |
