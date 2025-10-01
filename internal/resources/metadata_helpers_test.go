// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeepCopyMetadata(t *testing.T) {
	t.Run("nil map returns nil", func(t *testing.T) {
		result := DeepCopyMetadata(nil)
		assert.Nil(t, result)
	})

	t.Run("simple map is copied", func(t *testing.T) {
		src := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}
		dst := DeepCopyMetadata(src)

		assert.Equal(t, src, dst)

		// Modify destination, source should be unchanged
		dst["key1"] = "modified"
		assert.Equal(t, "value1", src["key1"])
	})

	t.Run("nested maps are deep copied", func(t *testing.T) {
		src := map[string]interface{}{
			"nested": map[string]interface{}{
				"inner": "value",
			},
		}
		dst := DeepCopyMetadata(src)

		// Modify nested map in destination
		nested := dst["nested"].(map[string]interface{})
		nested["inner"] = "modified"

		// Source should be unchanged
		srcNested := src["nested"].(map[string]interface{})
		assert.Equal(t, "value", srcNested["inner"])
	})
}

func TestDeepCopyStringMap(t *testing.T) {
	t.Run("nil map returns nil", func(t *testing.T) {
		result := DeepCopyStringMap(nil)
		assert.Nil(t, result)
	})

	t.Run("map is copied and independent", func(t *testing.T) {
		src := map[string]string{
			"key1": "value1",
			"key2": "value2",
		}
		dst := DeepCopyStringMap(src)

		assert.Equal(t, src, dst)

		// Modify destination
		dst["key1"] = "modified"

		// Source should be unchanged
		assert.Equal(t, "value1", src["key1"])
	})
}

func TestMergeMetadata(t *testing.T) {
	t.Run("both nil returns nil", func(t *testing.T) {
		result := MergeMetadata(nil, nil)
		assert.Nil(t, result)
	})

	t.Run("dst nil, src has values", func(t *testing.T) {
		src := map[string]interface{}{"key": "value"}
		result := MergeMetadata(nil, src)

		assert.Equal(t, src, result)

		// Modify result, src should be unchanged (deep copy)
		result["key"] = "modified"
		assert.Equal(t, "value", src["key"])
	})

	t.Run("dst has values, src nil", func(t *testing.T) {
		dst := map[string]interface{}{"key": "value"}
		result := MergeMetadata(dst, nil)
		assert.Equal(t, dst, result)
	})

	t.Run("dst keys take precedence", func(t *testing.T) {
		dst := map[string]interface{}{
			"shared":   "dst-value",
			"dst-only": "value1",
		}
		src := map[string]interface{}{
			"shared":   "src-value",
			"src-only": "value2",
		}

		result := MergeMetadata(dst, src)

		// dst values should win for shared keys
		assert.Equal(t, "dst-value", result["shared"])
		// both unique keys should be present
		assert.Equal(t, "value1", result["dst-only"])
		assert.Equal(t, "value2", result["src-only"])

		// Original maps should be unchanged
		assert.Equal(t, "dst-value", dst["shared"])
		assert.Equal(t, "src-value", src["shared"])
	})

	t.Run("no aliasing between result and inputs", func(t *testing.T) {
		dst := map[string]interface{}{"key": "value"}
		src := map[string]interface{}{"other": "data"}

		result := MergeMetadata(dst, src)

		// Modify result
		result["key"] = "modified"
		result["other"] = "changed"

		// Originals should be unchanged
		assert.Equal(t, "value", dst["key"])
		assert.Equal(t, "data", src["other"])
	})
}

func TestMergeStringMap(t *testing.T) {
	t.Run("both nil returns nil", func(t *testing.T) {
		result := MergeStringMap(nil, nil)
		assert.Nil(t, result)
	})

	t.Run("dst keys take precedence", func(t *testing.T) {
		dst := map[string]string{
			"shared":   "dst-value",
			"dst-only": "value1",
		}
		src := map[string]string{
			"shared":   "src-value",
			"src-only": "value2",
		}

		result := MergeStringMap(dst, src)

		assert.Equal(t, "dst-value", result["shared"])
		assert.Equal(t, "value1", result["dst-only"])
		assert.Equal(t, "value2", result["src-only"])
	})

	t.Run("no aliasing", func(t *testing.T) {
		dst := map[string]string{"key": "value"}
		src := map[string]string{"other": "data"}

		result := MergeStringMap(dst, src)

		result["key"] = "modified"
		assert.Equal(t, "value", dst["key"])
	})
}
