package core

import (
	"log"
	// "encoding/json" // Will be needed for parsing definition
)

// BehaviorTree represents an entire behavior tree.
type BehaviorTree struct {
	ID         string // Unique ID for this instance of the tree, if needed
	Root       Node
	// Blackboard *Blackboard // Tree-specific data store
	lastStatus Status // Optional: store the last status of the root node
}

// NewBehaviorTree creates a new behavior tree with the given root node.
func NewBehaviorTree(root Node, id string) *BehaviorTree {
	return &BehaviorTree{
		ID:         id,
		Root:       root,
		// Blackboard: NewBlackboard(), // Initialize if using a blackboard
		lastStatus: Invalid,
	}
}

// Tick executes the behavior tree starting from the root node.
// It requires a target object that the tree will operate on.
func (bt *BehaviorTree) Tick(target interface{}, deltaTime float64) Status {
	if bt.Root == nil {
		log.Println("Error: Behavior tree has no root node.")
		return Invalid
	}

	context := &TickContext{
		Target:    target,
		// Blackboard: bt.Blackboard,
		DeltaTime: deltaTime,
	}

	// log.Printf("Ticking Behavior Tree (ID: %s) for target: %+v", bt.ID, target)
	status := bt.Root.Tick(context)
	bt.lastStatus = status
	// log.Printf("Behavior Tree (ID: %s) Tick returned: %s", bt.ID, status)
	return status
}

// Halt sends a halt signal to the root of the tree, which should propagate down.
// This is used to stop any running nodes and reset their state.
func (bt *BehaviorTree) Halt(target interface{}) {
    if bt.Root == nil {
        return
    }
    context := &TickContext{
        Target: target,
        // Blackboard: bt.Blackboard,
    }
    log.Printf("Halting Behavior Tree (ID: %s)", bt.ID)
    bt.Root.Halt(context)
    bt.lastStatus = Invalid // Or some other appropriate status after halt
}


// Example of how one might load a tree from a definition later.
// This is a placeholder and non-functional for now.
/*
type NodeDefinition struct {
	Type     string            `json:"type"`
	Name     string            `json:"name,omitempty"`
	Params   map[string]string `json:"params,omitempty"` // For action/condition params
	Children []NodeDefinition  `json:"children,omitempty"` // For composite/decorator nodes
}

func (bt *BehaviorTree) LoadFromDefinition(definitionJSON string) error {
	var def NodeDefinition
	if err := json.Unmarshal([]byte(definitionJSON), &def); err != nil {
		return err
	}

	rootNode, err := parseNodeDefinition(def)
	if err != nil {
		return err
	}
	bt.Root = rootNode
	return nil
}

func parseNodeDefinition(def NodeDefinition) (Node, error) {
	// This function would recursively parse the definition and create nodes.
	// It would need a registry of node types and their constructors.
	// For example:
	// switch def.Type {
	// case "Sequence":
	//   seq := NewSequence()
	//   for _, childDef := range def.Children {
	//     childNode, err := parseNodeDefinition(childDef)
	//     if err != nil { return nil, err }
	//     seq.AddChild(childNode)
	//   }
	//   return seq, nil
	// case "LogAction":
	//   msg := def.Params["message"]
	//   statusStr := def.Params["status"] // needs parsing to Status type
	//   // ... parse statusStr ...
	//   return NewAction(def.Name, LogAction(def.Name, msg, parsedStatus)), nil
	// default:
	//   return nil, fmt.Errorf("unknown node type: %s", def.Type)
	// }
	return nil, fmt.Errorf("parsing from definition not yet implemented")
}
*/
