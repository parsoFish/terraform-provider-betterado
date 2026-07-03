package workitemtrackingprocess

import (
	"github.com/hashicorp/go-cty/cty/gocty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
)

// getBoolAttributeFromConfig returns a *bool from the configuration for the given key.
// Returns nil if the key is not set in the configuration, allowing
// distinction between an unset value and an explicit false.
func getBoolAttributeFromConfig(d *schema.ResourceData, key string) (*bool, error) {
	rawPlan := d.GetRawPlan()
	if !rawPlan.IsKnown() || rawPlan.IsNull() {
		return nil, nil
	}

	value := rawPlan.GetAttr(key)
	if !value.IsKnown() || value.IsNull() {
		return nil, nil
	}

	var val bool
	if err := gocty.FromCtyValue(value, &val); err != nil {
		return nil, err
	}
	return &val, nil
}

// findGroupById walks a FormLayout and returns the Group with the given ID (or nil).
func findGroupById(layout *workitemtrackingprocess.FormLayout, groupId string) *workitemtrackingprocess.Group {
	if layout == nil {
		return nil
	}
	pages := layout.Pages
	if pages == nil {
		return nil
	}
	for _, page := range *pages {
		group := findGroupInPage(&page, groupId)
		if group != nil {
			return group
		}
	}
	return nil
}

func findGroupInPage(page *workitemtrackingprocess.Page, groupId string) *workitemtrackingprocess.Group {
	sections := page.Sections
	if sections == nil {
		return nil
	}
	for _, section := range *sections {
		group := findGroupInSection(&section, groupId)
		if group != nil {
			return group
		}
	}
	return nil
}

func findGroupInSection(section *workitemtrackingprocess.Section, groupId string) *workitemtrackingprocess.Group {
	groups := section.Groups
	if groups == nil {
		return nil
	}
	for _, group := range *groups {
		id := group.Id
		if id == nil {
			continue
		}
		if *id == groupId {
			return &group
		}
	}
	return nil
}

// findControlInGroup returns the Control with the given ID from a Group (or nil).
func findControlInGroup(group *workitemtrackingprocess.Group, controlId string) *workitemtrackingprocess.Control {
	if group == nil || group.Controls == nil {
		return nil
	}
	for _, control := range *group.Controls {
		id := control.Id
		if id == nil {
			continue
		}
		if *id == controlId {
			c := control
			return &c
		}
	}
	return nil
}
