# Terraform Resource Scaffolder Skill

## Purpose

Generate boilerplate code for new Terraform resources following the patterns established in the official `terraform-provider-azuredevops` provider. This ensures consistency and reduces the manual work of setting up new resources.

## When to Use

- When adding a new Terraform resource to the provider
- When adding a new data source
- When you have a mapped API and need to create the Go implementation

## Workflow

### Step 1: Gather API Information

Before scaffolding, ensure you have:
1. The API reference doc in `docs/api-reference/` for the resource
2. A clear mapping of API fields → Terraform attributes
3. Understanding of which fields are Required, Optional, Computed, ForceNew
4. The CRUD endpoints and any special behaviors

### Step 2: Generate the Resource File

Create `azuredevops/internal/service/release/<subpackage>/resource_<name>.go` following this template:

```go
package <subpackage>

import (
    "context"
    "fmt"
    "strconv"

    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

    "github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Resource<Name> returns the Terraform resource schema
func Resource<Name>() *schema.Resource {
    return &schema.Resource{
        Description:   "<Human readable description>",
        CreateContext: resource<Name>Create,
        ReadContext:   resource<Name>Read,
        UpdateContext: resource<Name>Update,
        DeleteContext: resource<Name>Delete,

        Importer: &schema.ResourceImporter{
            StateContext: schema.ImportStatePassthroughContext,
        },

        Schema: map[string]*schema.Schema{
            "project_id": {
                Type:         schema.TypeString,
                Required:     true,
                ForceNew:     true,
                ValidateFunc: validation.IsUUID,
                Description:  "The ID of the project",
            },
            // ... add fields from API mapping
        },
    }
}

func resource<Name>Create(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    clients := m.(*client.AggregatedClient)
    projectID := d.Get("project_id").(string)

    // Expand terraform state to API object
    apiObj := expand<Name>(d)

    // Make API call
    created, err := clients.ReleaseClient.Create<Name>(ctx, projectID, apiObj)
    if err != nil {
        return diag.Errorf("creating <name>: %+v", err)
    }

    d.SetId(strconv.Itoa(created.ID))
    return resource<Name>Read(ctx, d, m)
}

func resource<Name>Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    clients := m.(*client.AggregatedClient)
    projectID := d.Get("project_id").(string)
    id, _ := strconv.Atoi(d.Id())

    apiObj, err := clients.ReleaseClient.Get<Name>(ctx, projectID, id)
    if err != nil {
        if isNotFound(err) {
            d.SetId("")
            return nil
        }
        return diag.Errorf("reading <name> (ID: %d): %+v", id, err)
    }

    flatten<Name>(d, apiObj)
    return nil
}

func resource<Name>Update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    clients := m.(*client.AggregatedClient)
    projectID := d.Get("project_id").(string)
    id, _ := strconv.Atoi(d.Id())

    apiObj := expand<Name>(d)
    apiObj.ID = id
    apiObj.Revision = d.Get("revision").(int)

    _, err := clients.ReleaseClient.Update<Name>(ctx, projectID, id, apiObj)
    if err != nil {
        return diag.Errorf("updating <name> (ID: %d): %+v", id, err)
    }

    return resource<Name>Read(ctx, d, m)
}

func resource<Name>Delete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    clients := m.(*client.AggregatedClient)
    projectID := d.Get("project_id").(string)
    id, _ := strconv.Atoi(d.Id())

    err := clients.ReleaseClient.Delete<Name>(ctx, projectID, id)
    if err != nil {
        return diag.Errorf("deleting <name> (ID: %d): %+v", id, err)
    }

    return nil
}
```

### Step 3: Generate Expand/Flatten Helpers

Create helper functions that convert between Terraform state and API objects. For nested structures, create separate expand/flatten functions for each level:

```go
// expand<Name> converts Terraform state to API object
func expand<Name>(d *schema.ResourceData) *api.<Name> {
    obj := &api.<Name>{
        Name: d.Get("name").(string),
    }
    // ... expand each field
    return obj
}

// flatten<Name> converts API object to Terraform state
func flatten<Name>(d *schema.ResourceData, obj *api.<Name>) {
    d.Set("name", obj.Name)
    // ... flatten each field
}

// For nested blocks:
func expandEnvironments(input []interface{}) []api.Environment {
    envs := make([]api.Environment, len(input))
    for i, raw := range input {
        v := raw.(map[string]interface{})
        envs[i] = api.Environment{
            Name: v["name"].(string),
            Rank: v["rank"].(int),
            // ...
        }
    }
    return envs
}

func flattenEnvironments(envs []api.Environment) []map[string]interface{} {
    result := make([]map[string]interface{}, len(envs))
    for i, env := range envs {
        result[i] = map[string]interface{}{
            "name": env.Name,
            "rank": env.Rank,
            // ...
        }
    }
    return result
}
```

### Step 4: Generate Test File

Create `resource_<name>_test.go` with:
- Basic acceptance test (create + verify)
- Complete acceptance test (all fields)
- Update test (modify + verify)
- Import test (import state)
- Destroy check function

### Step 5: Generate Example Configuration

Create `examples/<name>/main.tf` with example Terraform usage.

### Step 6: Generate Documentation

Create `docs/resources/<name>.md` with:
- Description
- Example usage
- Argument reference (all attributes)
- Attribute reference (computed fields)
- Import syntax

## Nested Block Schema Rules

For the release definition's deeply nested structure, follow these rules:

1. **Single nested object** → `TypeList` with `MaxItems: 1`
2. **List of nested objects** → `TypeList` of `schema.Resource`
3. **Map of simple values** → `TypeMap` with elem type
4. **Map of complex objects** → `TypeList` of blocks with `name` + `value` pattern
5. **Sets (unordered)** → `TypeSet` with appropriate hash function

Example for environments:
```go
"environment": {
    Type:     schema.TypeList,
    Required: true,
    MinItems: 1,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            "name":        {Type: schema.TypeString, Required: true},
            "rank":        {Type: schema.TypeInt, Required: true},
            "description": {Type: schema.TypeString, Optional: true},
            "pre_deploy_approvals": {
                Type:     schema.TypeList,
                Optional: true,
                MaxItems: 1,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{ ... },
                },
            },
            // ...
        },
    },
}
```
