# Output

The following output shows how specific offers are matched with stores:

- **Object offer_123 matched with stores:** [store_1, store_2, store_3]  
  Offer `offer_123` is associated with three stores.

- **Object offer_456 matched with stores:** [store_3]  
  Offer `offer_456` is associated with one store.

---

# Tree Structure

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