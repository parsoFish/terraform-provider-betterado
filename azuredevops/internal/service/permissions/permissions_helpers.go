package permissions

// permissions_helpers.go holds shared helper functions used by the framework
// implementations of the migrated permissions resources.

import (
	"context"
	"fmt"
	"strings"

	"github.com/ahmetb/go-linq"
	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// transformPath converts a Windows-style backslash path to a forward-slash path
// and strips leading/trailing slashes.
func transformPath(path string) string {
	paths := strings.Split(path, "\\")
	transformedPath := strings.Join(paths, "/")

	// remove slash at front of string
	transformedPath = strings.TrimPrefix(transformedPath, "/")

	// remove slash at end of string
	transformedPath = strings.TrimSuffix(transformedPath, "/")

	return transformedPath
}

// getQueryIDsFromPath resolves a path string to a list of query/folder IDs.
func getQueryIDsFromPath(ctx context.Context, wiqClient workitemtracking.Client, projectID string, path string) (*[]string, error) {
	var pathItems []string
	var err error
	var qry *workitemtracking.QueryHierarchyItem
	ret := []string{}

	path = strings.TrimSpace(path)
	linq.From(strings.Split(path, "/")).
		Where(func(elem interface{}) bool {
			return len(elem.(string)) > 0
		}).
		ToSlice(&pathItems)

	qry, err = wiqClient.GetQuery(ctx, workitemtracking.GetQueryArgs{
		Project: &projectID,
		Query:   converter.String("Shared Queries"),
		Depth:   converter.Int(1),
	})
	if err != nil {
		return nil, err
	}
	ret = append(ret, qry.Id.String())
	if len(pathItems) > 0 {
		for _, v := range pathItems {
			if qry.Children == nil || len(*qry.Children) == 0 {
				return nil, fmt.Errorf("Unable to find query [%s] in folder [%s] because it has no children", v, converter.ToString(qry.Name, qry.Id.String()))
			}

			segUUID, ok := uuid.Parse(v)
			chldIdx := -1
			for idx, chldItem := range *qry.Children {
				if ok == nil && strings.EqualFold(segUUID.String(), chldItem.Id.String()) {
					chldIdx = idx
				} else if chldItem.Name != nil && strings.EqualFold(*chldItem.Name, v) {
					chldIdx = idx
				}
			}

			if chldIdx < 0 {
				return nil, fmt.Errorf("Unable to find query [%s] in folder [%s]", v, converter.ToString(qry.Name, qry.Id.String()))
			}

			qry, err = wiqClient.GetQuery(ctx, workitemtracking.GetQueryArgs{
				Project: &projectID,
				Query:   converter.String((*qry.Children)[chldIdx].Id.String()),
				Depth:   converter.Int(1),
			})
			if err != nil {
				return nil, err
			}
			ret = append(ret, qry.Id.String())
		}
	}
	return &ret, nil
}
