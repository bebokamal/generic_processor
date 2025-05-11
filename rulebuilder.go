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
	values := []interface{}{"*"}
	negative := false

	attrDataRaw, ok := rule["attributes"].(Dict)[attrKey]
	if ok && attrDataRaw != nil {
		attrData := attrDataRaw.(Dict)
		values = attrData["values"].([]interface{})
		negative = attrData["negative"].(bool)
	}

	children := node["children"].(Dict)

	if negative {
		if _, exists := children["__not__"]; !exists {
			children["__not__"] = []Dict{}
		}
		notList := children["__not__"].([]Dict)
		nextNode := newNode("*", attrKey)
		children["__not__"] = append(notList, Dict{
			"excluded_values": values,
			"node":            nextNode,
		})
		insertRuleDict(nextNode, rule, attrs, level+1)
	} else {
		for _, val := range values {
			valStr := fmt.Sprintf("%v", val)
			if _, exists := children[valStr]; !exists {
				children[valStr] = newNode(valStr, attrKey)
			}
			insertRuleDict(children[valStr].(Dict), rule, attrs, level+1)
		}
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

func match(node Dict, object Dict, attrs []string, level int) []string {
	if codes, exists := node["codes"]; exists && len(codes.([]string)) > 0 {
		return codes.([]string)
	}

	var attrKey string
	for _, child := range node["children"].(Dict) {
		if childNode, ok := child.(Dict); ok {
			attrKey = childNode["attribute"].(string)
			break
		}
	}

	attrValsRaw := object["attributes"].(Dict)[attrKey]
	var values []interface{}
	if attrValsRaw == nil {
		values = []interface{}{"*"}
	} else {
		values = attrValsRaw.([]interface{})
	}

	matchedCodes := []string{}
	children := node["children"].(Dict)

	for _, val := range values {
		valStr := fmt.Sprintf("%v", val)
		if childNode, exists := children[valStr]; exists {
			matchedCodes = append(matchedCodes, match(childNode.(Dict), object, attrs, level+1)...)
		}
	}

	// Wildcard
	if wildcardNode, exists := children["*"]; exists {
		matchedCodes = append(matchedCodes, match(wildcardNode.(Dict), object, attrs, level+1)...)
	}

	// Negative checks
	if notNodesRaw, exists := children["__not__"]; exists {
		notNodes := notNodesRaw.([]Dict)
		for _, notEntry := range notNodes {
			excluded := notEntry["excluded_values"].([]interface{})
			excludedSet := map[string]bool{}
			for _, ex := range excluded {
				excludedSet[fmt.Sprintf("%v", ex)] = true
			}

			exclude := false
			for _, val := range values {
				if excludedSet[fmt.Sprintf("%v", val)] {
					exclude = true
					break
				}
			}

			if !exclude {
				nextNode := notEntry["node"].(Dict)
				matchedCodes = append(matchedCodes, match(nextNode, object, attrs, level+1)...)
			}
		}
	}

	return matchedCodes
}

func matchObjects(tree Dict, objects []Dict, attrs []string) map[string][]string {
	result := make(map[string][]string)
	for _, obj := range objects {
		matches := match(tree, obj, attrs, 0)
		objID := fmt.Sprintf("%v", obj["ID"])
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
	fmt.Println("tree", string(jsonBytes))
}

func main() {
	rules := []Dict{
		{
			"code": "store_1",
			"attributes": Dict{
				"country": Dict{"values": []interface{}{"US"}, "negative": false},
				"brand":   Dict{"values": []interface{}{"Nike"}, "negative": false},
			},
		},
		{
			"code": "store_2",
			"attributes": Dict{
				"country": Dict{"values": []interface{}{"US"}, "negative": false},
			},
		},
		{
			"code": "store_3",
			"attributes": Dict{
				"brand": Dict{"values": []interface{}{"Adidas", "Puma"}, "negative": true},
			},
		},
	}

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
				"brand":   []interface{}{"Reebok"},
			},
		},
	}

	attrs := []string{"country", "brand"}
	tree := BuildTreeFromRules(rules, attrs)
	matchedResults := matchObjects(tree, objects, attrs)

	for objID, matches := range matchedResults {
		fmt.Printf("Object %v matched with stores: %v\n", objID, matches)
	}

	PrintTree(tree)
}
