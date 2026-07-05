package workitemtrackingprocess

import (
	"context"
	"fmt"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
)

// pollUntilConsistent polls the refresh function until it returns "consistent"
// (or an error), with a minimum interval between polls and an overall timeout.
// It is a pure-Go polling helper with no third-party retry library dependency.
//
// refresh must return (result, state, err). "consistent" is the target state;
// any other non-error state is treated as pending. The function returns the
// final result on success, or an error on timeout / refresh error.
//
// continuousRequired is the number of consecutive "consistent" results needed
// before the poll is considered done (mirrors ContinuousTargetOccurence).
func pollUntilConsistent(ctx context.Context, timeout, minInterval time.Duration, continuousRequired int, refresh func() (interface{}, string, error)) (interface{}, error) {
	if continuousRequired <= 0 {
		continuousRequired = 1
	}
	deadline := time.Now().Add(timeout)
	consecutive := 0
	var lastResult interface{}
	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out waiting for consistent state after %s", timeout)
		}
		result, state, err := refresh()
		if err != nil {
			return nil, err
		}
		if state == "consistent" {
			consecutive++
			lastResult = result
			if consecutive >= continuousRequired {
				return lastResult, nil
			}
		} else {
			consecutive = 0
		}
		// Wait before next poll, but also honour context cancellation.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(minInterval):
		}
	}
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
