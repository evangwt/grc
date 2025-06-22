package models

import (
	"time"

	"gorm.io/gorm"
)

// BehaviorTree represents the main structure of a behavior tree.
type BehaviorTree struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `json:"name" gorm:"uniqueIndex"` // Enforce unique names for trees
	Definition string   `json:"definition" gorm:"type:TEXT"` // JSON string defining the tree structure (nodes and connections)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// RootNodeID *uint     `json:"root_node_id"` // Optional: If we want a direct link to a root node persisted in the Nodes table
	// Nodes      []Node    `json:"nodes" gorm:"foreignKey:TreeID"` // If nodes are stored separately and linked
}

// Node represents a single node within a behavior tree.
// For now, we'll store the tree definition as a JSON blob in BehaviorTree.Definition.
// This Node model can be used if we decide to store each node as a separate record.
// Or, it can be used for representing node types/templates.
// For simplicity in this step, we will not auto-migrate this Node model yet,
// as the primary way to store tree structure is via BehaviorTree.Definition.
// This model is here as a placeholder for future expansion if needed (e.g., for node templates).
type Node struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	// TreeID      *uint     `json:"tree_id"` // Foreign key to BehaviorTree if nodes are stored per tree
	Type        string    `json:"type"`    // e.g., "Sequence", "Selector", "Action", "Condition"
	Name        string    `json:"name"`    // User-defined name for the node instance / template name
	Description string    `json:"description"` // Description of what the node does (especially for templates)
	Parameters  string    `json:"parameters" gorm:"type:TEXT"` // JSON string for node-specific parameters (e.g., script for an action, default values for a template)
	IsTemplate  bool      `json:"is_template" gorm:"default:false"` // True if this is a node template
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ExecutionLog will store records of behavior tree executions.
type ExecutionLog struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	TreeID     uint      `json:"tree_id"` // ID of the BehaviorTree being executed
	RunID      string    `json:"run_id" gorm:"index"` // To group logs for a single execution run
	NodeID     string    `json:"node_id,omitempty"` // Client-side ID of the node this log entry pertains to (from the Definition JSON)
	NodeName   string    `json:"node_name,omitempty"` // Name of the node
	NodeType   string    `json:"node_type,omitempty"` // Type of the node
	Timestamp  time.Time `json:"timestamp" gorm:"index"`
	Status     string    `json:"status"` // e.g., "SUCCESS", "FAILURE", "RUNNING", "ERROR", "SKIPPED"
	Message    string    `json:"message" gorm:"type:TEXT"`
	Context    string    `json:"context,omitempty" gorm:"type:TEXT"` // JSON string of the object's state or relevant context at this point
}

// Global GORM database instance, to be initialized in main.go
var DB *gorm.DB
