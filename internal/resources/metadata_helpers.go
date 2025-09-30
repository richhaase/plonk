// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package resources

// DeepCopyMetadata creates a deep copy of a metadata map to avoid aliasing issues
func DeepCopyMetadata(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}

	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		// For nested maps, recursively copy
		if nestedMap, ok := v.(map[string]interface{}); ok {
			dst[k] = DeepCopyMetadata(nestedMap)
		} else {
			// For other types, direct copy is sufficient
			// Note: This doesn't deep-copy slices/pointers, but that's OK for our use case
			dst[k] = v
		}
	}
	return dst
}

// DeepCopyStringMap creates a deep copy of a string map
func DeepCopyStringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}

	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// MergeMetadata merges source metadata into destination without aliasing
// Keys in dst take precedence (won't be overwritten)
func MergeMetadata(dst, src map[string]interface{}) map[string]interface{} {
	if dst == nil && src == nil {
		return nil
	}
	if dst == nil {
		return DeepCopyMetadata(src)
	}
	if src == nil {
		return dst
	}

	// Create a new map to avoid mutating dst
	result := DeepCopyMetadata(dst)

	// Merge in values from src that don't exist in dst
	for k, v := range src {
		if _, exists := result[k]; !exists {
			if nestedMap, ok := v.(map[string]interface{}); ok {
				result[k] = DeepCopyMetadata(nestedMap)
			} else {
				result[k] = v
			}
		}
	}
	return result
}

// MergeStringMap merges source string map into destination
// Keys in dst take precedence (won't be overwritten)
func MergeStringMap(dst, src map[string]string) map[string]string {
	if dst == nil && src == nil {
		return nil
	}
	if dst == nil {
		return DeepCopyStringMap(src)
	}
	if src == nil {
		return dst
	}

	// Create new map to avoid mutating dst
	result := DeepCopyStringMap(dst)

	// Merge in values from src that don't exist in dst
	for k, v := range src {
		if _, exists := result[k]; !exists {
			result[k] = v
		}
	}
	return result
}
