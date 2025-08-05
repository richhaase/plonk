#!/usr/bin/env bash

# Assert output contains all of the provided strings
assert_output_contains_all() {
  for expected in "$@"; do
    assert_output --partial "$expected"
  done
}

# Assert output contains none of the provided strings
assert_output_contains_none() {
  for unexpected in "$@"; do
    refute_output --partial "$unexpected"
  done
}

# Assert JSON field has expected value
assert_json_field() {
  local field="$1"
  local expected="$2"
  local actual=$(echo "$output" | jq -r "$field" 2>/dev/null)

  if [[ "$actual" != "$expected" ]]; then
    echo "Expected $field to be '$expected', got '$actual'" >&2
    return 1
  fi
}

# Assert JSON field exists
assert_json_field_exists() {
  local field="$1"
  local actual=$(echo "$output" | jq -r "$field" 2>/dev/null)

  if [[ "$actual" == "null" ]] || [[ -z "$actual" ]]; then
    echo "Expected $field to exist in JSON output" >&2
    return 1
  fi
}

# Assert table contains row with values
assert_table_row() {
  local row_pattern="$1"
  if ! echo "$output" | grep -E "$row_pattern" > /dev/null; then
    echo "Expected table to contain row matching: $row_pattern" >&2
    return 1
  fi
}

# Assert exit code
assert_exit_code() {
  local expected="$1"
  if [[ "$status" -ne "$expected" ]]; then
    echo "Expected exit code $expected, got $status" >&2
    return 1
  fi
}

# Assert file exists
assert_file_exists() {
  local file="$1"
  if [[ ! -f "$file" ]]; then
    echo "Expected file to exist: $file" >&2
    return 1
  fi
}

# Assert directory exists
assert_dir_exists() {
  local dir="$1"
  if [[ ! -d "$dir" ]]; then
    echo "Expected directory to exist: $dir" >&2
    return 1
  fi
}
