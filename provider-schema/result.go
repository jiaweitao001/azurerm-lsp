package provider_schema

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Azure/ms-terraform-lsp/provider-schema/azurerm/schema"
	"github.com/Azure/ms-terraform-lsp/provider-schema/processors"
)

var finalTerraformObject processors.TerraformObjects

func GetFinalTerraformObject(objName string, isDataSource bool) *processors.TerraformObject {
	if finalTerraformObject == nil {
		terraformObject, err := processors.LoadProcessedOutput()
		if err != nil {
			panic(err)
		}

		finalTerraformObject = terraformObject
	}

	if isDataSource {
		objName = schema.InputDataSourcePrefix + objName
	}

	obj, _ := finalTerraformObject[objName]
	return obj
}

func GetFinalTerraformObjects() processors.TerraformObjects {
	if finalTerraformObject == nil {
		terraformObject, err := processors.LoadProcessedOutput()
		if err != nil {
			panic(err)
		}

		finalTerraformObject = terraformObject
	}

	return finalTerraformObject
}

func ListAllResourcesAndDataSources() []*processors.TerraformObject {
	resources := make([]*processors.TerraformObject, 0)

	for _, terraformObject := range GetFinalTerraformObjects() {
		resources = append(resources, terraformObject)
	}

	return resources
}

// NavigateToNestedBlock navigates down a nested block path and returns the final block or property
func NavigateToNestedBlock(objName, path string, isDataSource bool) (*schema.SchemaAttribute, error) {
	resource := GetFinalTerraformObject(objName, isDataSource)
	if resource == nil {
		return nil, fmt.Errorf("resource/data source '%s' not found", objName)
	}

	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid path '%s'", path)
	}

	curFields := resource.Fields
	var result *schema.SchemaAttribute
	var exists bool
	for _, part := range parts {
		result, exists = curFields[part]
		if !exists {
			return nil, fmt.Errorf("path '%s' not found in resource/data source '%s'", path, objName)
		}
		curFields = result.Fields
	}

	if result == nil {
		return nil, fmt.Errorf("path '%s' is nil in resource/data source '%s'", path, objName)
	}

	return result, nil
}

func GetObjectInfo(objName string, isDataSource bool) (*processors.TerraformObject, error) {
	resource := GetFinalTerraformObject(objName, isDataSource)
	if resource == nil {
		return nil, fmt.Errorf("resource/data source '%s' not found", objName)
	}

	return resource, nil
}

func GetPropertyInfo(objName, propertyPath string, isDataSource bool) (*schema.SchemaAttribute, error) {
	return NavigateToNestedBlock(objName, propertyPath, isDataSource)
}

func GetPossibleValuesForProperty(objName, propertyName string, isDataSource bool) ([]string, error) {
	block, err := NavigateToNestedBlock(objName, propertyName, isDataSource)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, fmt.Errorf("block '%s' not found in resource/data source '%s'", propertyName, objName)
	}

	if block.PossibleValues == nil {
		return nil, fmt.Errorf("no possible values found for block '%s' in resource/data source '%s'", propertyName, objName)
	}

	return block.GetAutoCompletePossibleValues(), nil
}

func ListDirectProperties(objName string, path string, isDataSource bool) ([]*schema.SchemaAttribute, error) {
	resource := GetFinalTerraformObject(objName, isDataSource)
	if resource == nil {
		return nil, fmt.Errorf("resource/data source '%s' not found", objName)
	}
	fields := resource.Fields

	if path != "" {
		block, err := NavigateToNestedBlock(objName, path, isDataSource)
		if err != nil {
			return nil, err
		}

		if block == nil {
			return nil, fmt.Errorf("block '%s' not found in resource/data source '%s'", path, objName)
		}

		fields = block.Fields
	}

	var properties []*schema.SchemaAttribute
	for _, property := range fields {
		if property.Computed {
			continue
		}

		properties = append(properties, property)
	}

	setSort(properties)

	return properties, nil
}

func setSort(properties []*schema.SchemaAttribute) {
	slices.SortFunc(properties, func(a, b *schema.SchemaAttribute) int {
		if a.Required && !b.Required {
			return -1
		}
		if !a.Required && b.Required {
			return 1
		}
		return strings.Compare(a.Name, b.Name)
	})

	sortIndex := 0
	for _, property := range properties {
		property.SetSortOrder(fmt.Sprintf("%d", sortIndex))
		sortIndex++
	}
}

func GetSnippet(objName string, isDataSource bool) (string, error) {
	resource := GetFinalTerraformObject(objName, isDataSource)
	if resource == nil {
		return "", fmt.Errorf("resource/data source '%s' not found", objName)
	}

	return resource.GetSnippet(), nil
}

func GetPropertyDocContent(objName string, property *schema.SchemaAttribute, isDataSource bool) string {
	if property == nil {
		return ""
	}

	propertyDescription := property.GetDescription()

	// try to get direct properties of this property
	directProperties, err := ListDirectProperties(objName, property.AttributePath, isDataSource)
	if err != nil || len(directProperties) == 0 {
		return propertyDescription
	}

	// if there are direct properties, append them to the description
	var directPropertiesDescriptions []string
	for _, directProperty := range directProperties {
		directPropertiesDescriptions = append(directPropertiesDescriptions, fmt.Sprintf(" - %s: %s", directProperty.Name, directProperty.GetDescription()))
	}

	return fmt.Sprintf("%s\n\n%s", propertyDescription, strings.Join(directPropertiesDescriptions, "\n"))
}

func GetResourceContent(objName string, isDataSource bool) (string, bool, error) {
	resourceInfo, err := GetObjectInfo(objName, isDataSource)
	if err != nil {
		return "", false, fmt.Errorf("error retrieving resource info: %v", err)
	}
	return fmt.Sprintf(ResourceTemplate,
		objName,
		resourceInfo.GetResourceOrDataSourceDocLink(),
		resourceInfo.GetGitHubIssueLink(),
		resourceInfo.GetRaiseGitHubIssueLink(),
		resourceInfo.GetDocContent(),
	), resourceInfo.IsDataSource(), nil
}

func GetAttributeContent(objName, path string, isDataSource bool) (string, *schema.SchemaAttribute, error) {
	obj, err := GetObjectInfo(objName, isDataSource)
	if err != nil {
		return "", nil, fmt.Errorf("error retrieving object info: %v", err)
	}
	prop, err := GetPropertyInfo(objName, path, isDataSource)
	if err != nil {
		return "", nil, fmt.Errorf("error retrieving property info: %v", err)
	}
	return fmt.Sprintf(AttributeTemplate,
		prop.Name,
		prop.GetRequirementType(),
		prop.AttributeType.FriendlyName(),
		prop.GetAttributeDocLink(obj.GetResourceOrDataSourceDocLink()),
		prop.GetGitHubIssueLink(),
		prop.GetRaiseGitHubIssueLink(),
		strings.Join(prop.GetDetails(), "\n"),
		GetPropertyDocContent(objName, prop, isDataSource),
	), prop, nil
}
