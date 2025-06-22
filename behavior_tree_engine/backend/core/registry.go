package core

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/dop251/goja" // Import Goja
)

// NodeConstructor is a function that creates a Node.
type NodeConstructor func(params map[string]interface{}, children []Node) (Node, error)

var nodeRegistry = make(map[string]NodeConstructor)

// RegisterNode registers a node type with its constructor.
func RegisterNode(typeName string, constructor NodeConstructor) {
	if _, exists := nodeRegistry[typeName]; exists {
		log.Printf("Warning: Node type '%s' is being re-registered.", typeName)
	}
	nodeRegistry[typeName] = constructor
}

// NodeDefinition as stored in JSON
type NodeDefinition struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"` // VueFlow node type
	Label    string                 `json:"label"`
	Position XYPosition             `json:"position"`
	Data     map[string]interface{} `json:"data"` // Includes 'engineType', 'script', etc.
}

type EdgeDefinition struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type TreeDefinition struct {
	Nodes []NodeDefinition `json:"nodes"`
	Edges []EdgeDefinition `json:"edges"`
}

type XYPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// InitDefaultRegistry registers the standard node types.
func InitDefaultRegistry() {
	RegisterNode("sequence", func(params map[string]interface{}, children []Node) (Node, error) {
		seq := NewSequence()
		if label, ok := params["label"].(string); ok {
			seq.Name = label
		} else {
			seq.Name = "Sequence"
		}
		seq.AddChildren(children...)
		return seq, nil
	})

	RegisterNode("selector", func(params map[string]interface{}, children []Node) (Node, error) {
		sel := NewSelector()
		if label, ok := params["label"].(string); ok {
			sel.Name = label
		} else {
			sel.Name = "Selector"
		}
		sel.AddChildren(children...)
		return sel, nil
	})

	RegisterNode("action", func(params map[string]interface{}, children []Node) (Node, error) {
		script, _ := params["script"].(string)
		label, ok := params["label"].(string)
		if !ok {
			label = "Unnamed Action"
		}

		actionFunc := func(ctx *TickContext) Status {
			vm := goja.New()
			jsTarget := make(map[string]interface{}) // Default to empty if target is nil or wrong type

			if ctx.Target != nil {
				if targetMap, ok := ctx.Target.(map[string]interface{}); ok {
					// Create a new map to avoid modifying the original TargetState map directly by reference
					// if Goja's Set function or script execution modifies it internally.
					for k, v := range targetMap {
						jsTarget[k] = v
					}
				} else {
					log.Printf("Warning: Action '%s' target is not a map[string]interface{}, type: %T. Script will receive an empty target.", label, ctx.Target)
				}
			}
			vm.Set("target", jsTarget)

			vm.Set("log", func(call goja.FunctionCall) goja.Value {
				var args []interface{}
				for _, arg := range call.Arguments {
					args = append(args, arg.Export())
				}
				log.Printf("ScriptLog (Action: %s): %s", label, fmt.Sprint(args...))
				return vm.ToValue(nil)
			})

			_, err := vm.RunString(script)
			if err != nil {
				log.Printf("Error executing Action script for '%s': %v. Script: \n%s", label, err, script)
				return Failure
			}

			// Update the original context's TargetState with changes from jsTarget
            // This makes modifications by the script visible to the Go side.
            if currentTargetMap, ok := ctx.Target.(map[string]interface{}); ok {
                for k, v := range jsTarget {
                    currentTargetMap[k] = v
                }
            }

			return Success
		}
		return NewAction(label, actionFunc), nil
	})

	RegisterNode("condition", func(params map[string]interface{}, children []Node) (Node, error) {
		script, _ := params["script"].(string)
		label, ok := params["label"].(string)
		if !ok {
			label = "Unnamed Condition"
		}

		conditionFunc := func(ctx *TickContext) Status {
			vm := goja.New()
			// For conditions, target is usually read-only, so direct exposure is fine.
			if ctx.Target != nil {
				vm.Set("target", ctx.Target)
			} else {
				vm.Set("target", make(map[string]interface{}))
			}

			vm.Set("log", func(call goja.FunctionCall) goja.Value {
				var args []interface{}
				for _, arg := range call.Arguments {
					args = append(args, arg.Export())
				}
				log.Printf("ScriptLog (Condition: %s): %s", label, fmt.Sprint(args...))
				return vm.ToValue(nil)
			})

			result, err := vm.RunString(script)
			if err != nil {
				log.Printf("Error executing Condition script for '%s': %v. Script: \n%s", label, err, script)
				return Failure
			}

			if goja.IsTruthy(result) {
				return Success
			}
			return Failure
		}
		return NewAction(label, conditionFunc), nil // Using ActionNode for conditions
	})
}

// ParseTreeDefinition remains largely the same
func ParseTreeDefinition(jsonDefinition string) (*BehaviorTree, error) {
	var def TreeDefinition
	if err := json.Unmarshal([]byte(jsonDefinition), &def); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tree definition: %w", err)
	}

	if len(def.Nodes) == 0 {
		// Allow empty trees to be parsed, they just won't do anything.
		// Or, return an error if a root node is strictly required.
		// For now, let's create a BT with a nil root, which Tick will handle.
		log.Printf("Warning: Parsing a tree definition with no nodes.")
        bt := NewBehaviorTree(nil, "empty_tree")
		return bt, nil
	}

	nodesMap := make(map[string]Node)
	nodeDefinitionsMap := make(map[string]*NodeDefinition)
	parentToChildrenGraph := make(map[string][]string)

	for _, nodeDef := range def.Nodes {
		nodeDefinitionsMap[nodeDef.ID] = &nodeDef
	}

	for _, edge := range def.Edges {
		parentToChildrenGraph[edge.Source] = append(parentToChildrenGraph[edge.Source], edge.Target)
	}

	isTarget := make(map[string]bool)
	for _, edge := range def.Edges {
		isTarget[edge.Target] = true
	}

	var rootNodeIDs []string
	for _, nodeDef := range def.Nodes {
		if !isTarget[nodeDef.ID] {
			rootNodeIDs = append(rootNodeIDs, nodeDef.ID)
		}
	}

	if len(rootNodeIDs) == 0 {
        if len(def.Nodes) == 1 { // Single node is always root
            rootNodeIDs = []string{def.Nodes[0].ID}
        } else if len(def.Nodes) > 1 { // More than one node but no clear root (e.g. cycle)
		    return nil, fmt.Errorf("no root node found in tree definition (a root node is not a target of any edge and there are multiple nodes)")
        } else { // len(def.Nodes) == 0, handled above
            // This case should not be reached if len(def.Nodes) == 0 is handled earlier
            bt := NewBehaviorTree(nil, "empty_tree_no_root_ids")
		    return bt, nil
        }
	}
	if len(rootNodeIDs) > 1 {
		log.Printf("Warning: Multiple root nodes found (%v). Using the first one: %s", rootNodeIDs, rootNodeIDs[0])
	}
	rootID := rootNodeIDs[0]

	var buildNode func(nodeID string) (Node, error)
	buildNode = func(nodeID string) (Node, error) {
		if node, exists := nodesMap[nodeID]; exists {
			return node, nil
		}

		nodeDef, ok := nodeDefinitionsMap[nodeID]
		if !ok {
			return nil, fmt.Errorf("node definition not found for ID: %s", nodeID)
		}

		engineType, ok := nodeDef.Data["engineType"].(string)
		if !ok {
			// Attempt to infer engineType for default VueFlow nodes if they are sequence/selector
			if nodeDef.Type == "default" || nodeDef.Type == "" { // "" can be from older VueFlow versions or simple custom nodes
				nodeLabelLower := strings.ToLower(nodeDef.Label)
				if strings.Contains(nodeLabelLower, "sequence") {
					engineType = "sequence"
				} else if strings.Contains(nodeLabelLower, "selector") {
					engineType = "selector"
				} else {
					return nil, fmt.Errorf("node ID %s (Label: %s, VueFlow Type: %s) missing 'engineType' in data and could not infer", nodeDef.ID, nodeDef.Label, nodeDef.Type)
				}
				log.Printf("Inferred engineType '%s' for node ID %s (Label: %s)", engineType, nodeDef.ID, nodeDef.Label)
			} else {
				return nil, fmt.Errorf("node ID %s (Label: %s, VueFlow Type: %s) missing 'engineType' in data", nodeDef.ID, nodeDef.Label, nodeDef.Type)
			}
		}

		constructor, exists := nodeRegistry[engineType]
		if !exists {
			return nil, fmt.Errorf("no constructor registered for node type: %s (engineType)", engineType)
		}

		var childNodes []Node
		childNodeIDs := parentToChildrenGraph[nodeID]
		for _, childID := range childNodeIDs {
			childNode, err := buildNode(childID)
			if err != nil {
				return nil, fmt.Errorf("failed to build child node %s for parent %s: %w", childID, nodeID, err)
			}
			childNodes = append(childNodes, childNode)
		}

		node, err := constructor(nodeDef.Data, childNodes)
		if err != nil {
			return nil, fmt.Errorf("failed to construct node ID %s (type %s): %w", nodeID, engineType, err)
		}

		nodesMap[nodeID] = node
		return node, nil
	}

	rootEngineNode, err := buildNode(rootID)
	if err != nil {
		return nil, fmt.Errorf("failed to build behavior tree from root %s: %w", rootID, err)
	}

	bt := NewBehaviorTree(rootEngineNode, "parsed_tree")
	return bt, nil
}

func init() {
	InitDefaultRegistry()
}
