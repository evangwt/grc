package core

// Status represents the execution status of a node.
type Status int

const (
	Invalid Status = iota // Initial status, or error
	Success
	Failure
	Running
)

// String returns a human-readable representation of the status.
func (s Status) String() string {
	switch s {
	case Success:
		return "SUCCESS"
	case Failure:
		return "FAILURE"
	case Running:
		return "RUNNING"
	default:
		return "INVALID"
	}
}

// TickContext holds contextual information for a node's tick.
// This will be expanded to include the target object, blackboard, etc.
type TickContext struct {
	Target interface{} // The object the Behavior Tree is operating on
	// Blackboard *Blackboard // A shared data store for the tree
	DeltaTime float64 // Time since last tick, useful for time-dependent actions
	// More fields can be added: tree instance id, debug logger, etc.
}

// Node is the interface that all behavior tree nodes must implement.
type Node interface {
	Tick(context *TickContext) Status
	// Halt is called when a higher priority node aborts the execution of this node
	// or when the tree itself is halted. It should reset the node's internal state.
	Halt(context *TickContext)
	// GetID returns a unique identifier for the node instance (if needed for logging/debugging).
	// GetID() string
	// GetType returns the type of the node (e.g., "Sequence", "Action<MoveTo>").
	// GetType() string
}

// BaseNode provides a common structure for all nodes.
// It can handle common properties like a name or ID.
type BaseNode struct {
	ID   string
	Name string
	Type string
	// Children would typically not be here if specific composites manage them.
	// However, some designs put a generic children slice here.
	// For this iteration, composite nodes will manage their own children.
}

// CompositeNode is a base for nodes that have children (e.g., Sequence, Selector).
type CompositeNode struct {
	BaseNode
	Children []Node
}

// AddChild adds a child node to the composite node.
func (cn *CompositeNode) AddChild(child Node) {
	cn.Children = append(cn.Children, child)
}

// AddChildren adds multiple child nodes to the composite node.
func (cn *CompositeNode) AddChildren(children ...Node) {
	cn.Children = append(cn.Children, children...)
}

// Halt default implementation for composite nodes: halt all children.
func (cn *CompositeNode) Halt(context *TickContext) {
	for _, child := range cn.Children {
		child.Halt(context)
	}
	// Specific composites might need to reset their own state (e.g., running child index)
}


// DecoratorNode is a base for nodes that have a single child.
type DecoratorNode struct {
	BaseNode
	Child Node
}

// SetChild sets the child for the decorator node.
func (dn *DecoratorNode) SetChild(child Node) {
	dn.Child = child
}

// Halt default implementation for decorator nodes: halt the child.
func (dn *DecoratorNode) Halt(context *TickContext) {
	if dn.Child != nil {
		dn.Child.Halt(context)
	}
	// Specific decorators might need to reset their own state
}

// LeafNode is a base for nodes that have no children (e.g., Action, Condition).
type LeafNode struct {
	BaseNode
}

// Halt default implementation for leaf nodes (often does nothing unless the node has internal state).
func (ln *LeafNode) Halt(context *TickContext) {
	// Reset internal state if any
}
