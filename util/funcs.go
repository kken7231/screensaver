// Package util provides utility functions for various operations.
package util

import "reflect"

// StructToMap converts a struct to a map, with field names as keys.
// It handles nested structs and slices of structs.
func StructToMap(data interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	v := reflect.ValueOf(data)

	// Dereference pointers to get to the underlying value
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Ensure the input is a struct
	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		jsonTag := fieldType.Tag.Get("json")

		// Use the field name if the JSON tag is empty
		fieldName := fieldType.Name
		if jsonTag != "" && jsonTag != "-" {
			fieldName = jsonTag
		}

		switch field.Kind() {
		case reflect.Struct:
			// Recursively convert nested structs to maps
			result[fieldName] = StructToMap(field.Interface())
		case reflect.Slice:
			// Convert slices to slices of interfaces
			sliceResult := make([]interface{}, field.Len())
			for j := 0; j < field.Len(); j++ {
				if field.Index(j).Kind() == reflect.Struct {
					sliceResult[j] = StructToMap(field.Index(j).Interface())
				} else {
					sliceResult[j] = field.Index(j).Interface()
				}
			}
			result[fieldName] = sliceResult
		default:
			// Add the field value to the map
			result[fieldName] = field.Interface()
		}
	}
	return result
}

// MergeMaps merges two maps into one. If there are duplicate keys, values from map2 overwrite values from map1.
func MergeMaps(map1, map2 map[string]interface{}) map[string]interface{} {
	mergedMap := make(map[string]interface{})

	// Copy map1 into mergedMap
	for key, value := range map1 {
		mergedMap[key] = value
	}

	// Copy map2 into mergedMap, overwriting any duplicate keys
	for key, value := range map2 {
		mergedMap[key] = value
	}

	return mergedMap
}
