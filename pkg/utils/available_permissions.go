package utils

import (
	"backend-v2/model"
	"encoding/json"
	"os"
)

func AvailablePermissionList() ([]model.Permission, error) {
	content, err := os.ReadFile("./configurations/permissions.json")
	if err != nil {
		return []model.Permission{}, err
	}

	var fileData []struct {
		ResourceName string `json:"resourceName"`
		ResourceURL  string `json:"resourceURL"`
	}

	err = json.Unmarshal(content, &fileData)
	if err != nil {
		return []model.Permission{}, err
	}

	var result []model.Permission
	for _, permission := range fileData {
		result = append(result, model.Permission{
			ID:           0,
			RoleID:       0,
			ResourceUrl:  permission.ResourceURL,
			ResourceName: permission.ResourceName,
			R:            false,
			W:            false,
			U:            false,
			D:            false,
		})
	}

  return result, nil
}
