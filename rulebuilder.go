package main

import (
	"encoding/json"
	"fmt"
)

type Dict = map[string]interface{}

func newNode(value string, attribute string) Dict {
	return Dict{
		"value":     value,
		"attribute": attribute,
		"children":  Dict{},
	}
}

func insertRuleDict(node Dict, rule Dict, attrs []string, level int) {
	if level == len(attrs) {
		// Leaf node
		code := fmt.Sprintf("%v", rule["code"])
		if _, exists := node["codes"]; !exists {
			node["codes"] = []string{}
		}
		node["codes"] = appendIfMissing(node["codes"].([]string), code)
		delete(node, "children")
		return
	}

	attrKey := attrs[level]
	attrValsRaw := rule["attributes"].(Dict)[attrKey]
	var values []interface{}
	if attrValsRaw == nil || len(attrValsRaw.([]interface{})) == 0 {
		values = []interface{}{"*"}
	} else {
		values = attrValsRaw.([]interface{})
	}

	for _, val := range values {
		valStr := fmt.Sprintf("%v", val)
		children := node["children"].(Dict)

		if _, exists := children[valStr]; !exists {
			children[valStr] = newNode(valStr, attrKey)
		}
		insertRuleDict(children[valStr].(Dict), rule, attrs, level+1)
	}
	if level == 0 {
		delete(node, "attribute")
	}
}

func BuildTreeFromRules(rules []Dict, attrs []string) Dict {
	root := newNode("*", "")
	for _, rule := range rules {
		insertRuleDict(root, rule, attrs, 0)
	}
	return root
}

func appendIfMissing(slice []string, item string) []string {
	for _, v := range slice {
		if v == item {
			return slice
		}
	}
	return append(slice, item)
}

// Match function that checks an object against the tree
func match(node Dict, object Dict, attrs []string, level int) []string {
	if codes, exists := node["codes"]; exists && len(codes.([]string)) > 0 {
		return codes.([]string)
	}
	var attrKey string
	// Access any child from the map (assuming there is at least one child)
	for _, child := range node["children"].(Dict) {
		// Assume the first valid child is a Dict
		firstChildNode, ok := child.(Dict)
		if ok {
			attrKey = firstChildNode["attribute"].(string)
			break
		}
	}
	attrValsRaw := object["attributes"].(Dict)[attrKey]
	var values []interface{}
	if attrValsRaw == nil || len(attrValsRaw.([]interface{})) == 0 {
		values = []interface{}{"*"}
	} else {
		values = attrValsRaw.([]interface{})
	}

	var matchedCodes []string
	children := node["children"].(Dict)

	for _, val := range values {
		valStr := fmt.Sprintf("%v", val)
		if childNode, exists := children[valStr]; exists {
			matchedCodes = append(matchedCodes, match(childNode.(Dict), object, attrs, level+1)...)
		}
	}

	// Check for wildcard path "*"
	if wildcardNode, exists := children["*"]; exists {
		matchedCodes = append(matchedCodes, match(wildcardNode.(Dict), object, attrs, level+1)...)
	}

	return matchedCodes
}

// Function to loop over multiple objects and send them to the main match function, returning a map of matches
func matchObjects(tree Dict, objects []Dict, attrs []string) map[string][]string {
	result := make(map[string][]string)

	for _, obj := range objects {
		matches := match(tree, obj, attrs, 0)
		objID := fmt.Sprintf("%v", obj["ID"]) // Assuming ID is a string in the object
		result[objID] = matches
	}

	return result
}
func PrintTree(tree Dict) {
	jsonBytes, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(string(jsonBytes))
}

func main() {
	rules := []Dict{
		{
			"code": "store_1",
			"attributes": Dict{
				"country": []interface{}{"US"},
				"brand":   []interface{}{"Nike"},
			},
		},
		{
			"code": "store_2",
			"attributes": Dict{
				"country": []interface{}{"US"},
			},
		},
	}

	// Objects to match
	objects := []Dict{
		{
			"ID": "offer_123",
			"attributes": Dict{
				"country": []interface{}{"US"},
				"brand":   []interface{}{"Nike"},
			},
		},
		{
			"ID": "offer_456",
			"attributes": Dict{
				"country": []interface{}{"UK"},
			},
		},
	}

	attrs := []string{"country", "brand"}

	// Build the tree from rules
	tree := BuildTreeFromRules(rules, attrs)

	// Match all objects and get the result as a map
	matchedResults := matchObjects(tree, objects, attrs)

	// Print the matched results
	for objID, matches := range matchedResults {
		fmt.Printf("Object %v matched with stores: %v\n", objID, matches)
	}

	// Print the tree as JSON
	PrintTree(tree)
}
