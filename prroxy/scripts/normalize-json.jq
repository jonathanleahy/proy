# Deep normalization for JSON comparison
# Recursively sorts object keys and array elements for deterministic comparison

def normalize:
  if type == "object" then
    # Sort object keys and recursively normalize values
    to_entries | map({key, value: (.value | normalize)}) | sort_by(.key) | from_entries
  elif type == "array" then
    # Sort array elements by their normalized JSON representation
    # This ensures consistent ordering regardless of input order
    map(normalize) | sort_by(tojson)
  else
    # Primitive values (string, number, boolean, null) pass through unchanged
    .
  end;

# Apply normalization and sort keys at top level
normalize | .
