package core

import "log"

// ActionFunc is a function type that can be executed by an ActionNode.
// It takes a TickContext and returns a Status.
type ActionFunc func(context *TickContext) Status

// ActionNode is a leaf node that executes a specific action.
// The action is defined by an ActionFunc.
type ActionNode struct {
	LeafNode
	action ActionFunc
}

// NewAction creates a new ActionNode with the given name and action function.
func NewAction(name string, action ActionFunc) *ActionNode {
	a := &ActionNode{
		action: action,
	}
	a.Name = name
	a.Type = "Action"
	return a
}

// Tick executes the action function.
func (a *ActionNode) Tick(context *TickContext) Status {
	if a.action == nil {
		log.Printf("Error: ActionNode '%s' has no action function defined.", a.Name)
		return Failure // Or Invalid
	}
	// log.Printf("Ticking ActionNode: %s", a.Name)
	return a.action(context)
}

// Halt for ActionNode typically does nothing unless the action itself needs explicit cleanup
// for long-running tasks. For simple actions, LeafNode's Halt is sufficient.
// If an action could be long-running and needs interruption, this method would handle it.
// For example, cancelling a network request or stopping a timer.
// func (a *ActionNode) Halt(context *TickContext) {
//    log.Printf("Halting ActionNode: %s", a.Name)
//    a.LeafNode.Halt(context) // Call base Halt
//    // Add specific halt logic here if needed
// }


// --- Example Simple Actions ---

// LogAction is a simple action that logs a message and returns a specified status.
func LogAction(name string, message string, statusToReturn Status) ActionFunc {
	return func(context *TickContext) Status {
		// Access target from context if needed:
		// if context.Target != nil {
		//	 log.Printf("[%s] Target: %+v, Message: %s", name, context.Target, message)
		// } else {
		//	 log.Printf("[%s] Message: %s", name, message)
		// }
		log.Printf("LogAction [%s]: %s. Returning %s.", name, message, statusToReturn)
		return statusToReturn
	}
}

// WaitAction is an example of an action that could be stateful and return Running.
// This is a simplified version; a real WaitAction would use DeltaTime from TickContext
// or a persistent timer.
func WaitAction(name string, duration float64 /*seconds*/) ActionFunc {
    // These would be part of the node's state if it were a struct method,
    // or managed via a blackboard/context for a func.
    // For this simple func, it's not truly stateful across ticks without external state.
    // var startTime time.Time
    // var started bool

	return func(context *TickContext) Status {
		log.Printf("WaitAction [%s]: Pretending to wait for %.2f seconds. Returning Success for now.", name, duration)
        // Proper implementation would involve:
        // if !started {
        //     startTime = time.Now()
        //     started = true
        //     return Running
        // }
        // if time.Since(startTime).Seconds() < duration {
        //     return Running
        // }
        // started = false // Reset for next time
		return Success
	}
}
