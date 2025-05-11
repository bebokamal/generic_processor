# Rule Tree Matching System

This document outlines the design and implementation of a rule tree matching system that supports both positive (inclusive) and negative (exclusive) rule matching across multiple attributes.

---

## Overview

The system builds a decision tree from a set of rules, where each rule specifies conditions on one or more attributes. Each attribute can define values that:

- Must be matched (positive rule)
- Must NOT be matched (negative rule)

The tree structure enables efficient matching of objects against rules.

---

## Rule Format

Each rule is a dictionary with a unique code and an attributes map:

```go
Dict{
  "code": "store_1",  // Unique identifier for the rule.
  "attributes": Dict{ // Map of attributes and their conditions.
    "brand": Dict{
      "values": []interface{}{...},  // List of values for the attribute.
      "negative": bool,              // Indicates if the rule is negative (true) or positive (false).
    },
    ...
  },
}
```

---

## Object Format

Each object to match is structured as:

```go
Dict{
  "ID": "object_1",  // Unique identifier for the object.
  "attributes": Dict{ // Map of attributes and their values.
    "brand": []interface{}{...},  // List of values for the attribute.
    ...
  },
}
```

Each attribute contains a list of values (usually of length 1).

---

## Tree Structure

### Positive Children

For positive rules, a node stores children under each possible value:

```go
"children": {
  "Nike": { ... },  // Node for the "Nike" value.
  "Adidas": { ... },  // Node for the "Adidas" value.
  "*": { ... }  // Wildcard support for any value.
}
```

### Negative Rules

Negative conditions are stored in a dedicated list under a `__not__` key:

```go
"children": {
  "__not__": [
    {
      "excluded_values": ["Puma", "Reebok"],  // Values to exclude.
      "node": { ... }  // Node to proceed to if value is NOT in excluded_values.
    },
    ...
  ]
}
```

Each entry means: if the object value is NOT in `excluded_values`, proceed to the node.

### Wildcard Matching

Wildcard values are represented by `"*"`. They match any value in the attribute.

---

## Matching Process

1. Start at the root of the tree.
2. For each attribute in order:
    - Check positive children: follow the path for any matching value.
    - Check `__not__` children: follow the path if the value is NOT in `excluded_values`.
3. Repeat until a leaf node is reached.
4. Collect all rule codes at the leaf.

---

## Time Complexity

Let:

- `R` = number of rules
- `A` = number of attributes
- `V` = values per attribute
- `N` = number of objects
- `M` = number of negative branches per node

### Build Phase

- Complexity: `O(R × A × V)`

### Match Phase

- Positive-only: `O(N × A)`
- With negatives: `O(N × A × M)` (M is usually small)

---

## Benefits

- Fast lookups due to map-based tree.
- Compact representation using one rule node, even when the condition applies across multiple branches.
- Wildcard and negative logic are natively supported.

---

## Future Extensions

- Support range-based conditions (e.g., `price > 100`).
- Priority ranking for matching rules.
- Fallback strategies when no match is found.

---

## Output

The following output shows how specific offers are matched with stores:

- **Object `offer_123` matched with stores:** `[store_1, store_2, store_3]`  
  Offer `offer_123` is associated with three stores.

- **Object `offer_456` matched with stores:** `[store_3]`  
  Offer `offer_456` is associated with one store.

---

## Tree Structure Example

The tree structure represents a hierarchy of attributes and their corresponding values:

```json
{
  "children": {
    "*": {
      "attribute": "country",
      "children": {
        "__not__": [
          {
            "excluded_values": ["Adidas", "Puma"],
            "node": {
              "attribute": "brand",
              "codes": ["store_3"],
              "value": "*"
            }
          }
        ]
      },
      "value": "*"
    },
    "US": {
      "attribute": "country",
      "children": {
        "*": {
          "attribute": "brand",
          "codes": ["store_2"],
          "value": "*"
        },
        "Nike": {
          "attribute": "brand",
          "codes": ["store_1"],
          "value": "Nike"
        }
      },
      "value": "US"
    }
  },
  "value": "*"
}
```