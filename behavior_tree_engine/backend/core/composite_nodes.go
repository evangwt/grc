package core

// --- Sequence Node ---

// SequenceNode ticks its children sequentially as long as they return Success.
// If any child returns Failure, the SequenceNode immediately returns Failure.
// If any child returns Running, the SequenceNode immediately returns Running, remembering the running child.
// If all children return Success, the SequenceNode returns Success.
type SequenceNode struct {
	CompositeNode
	runningChildIndex int // Index of the child that is currently running
}

// NewSequence creates a new SequenceNode.
func NewSequence(children ...Node) *SequenceNode {
	s := &SequenceNode{runningChildIndex: 0}
	s.Type = "Sequence"
	s.AddChildren(children...)
	return s
}

// Tick executes the children of the SequenceNode.
func (s *SequenceNode) Tick(context *TickContext) Status {
	for i := s.runningChildIndex; i < len(s.Children); i++ {
		child := s.Children[i]
		status := child.Tick(context)

		switch status {
		case Failure:
			s.Halt(context) // Halt all children and reset self before returning
			return Failure
		case Running:
			s.runningChildIndex = i // Remember the running child
			return Running
		case Success:
			// Continue to the next child
		default: // Invalid status
			s.Halt(context)
			return Failure // Or some error status if defined
		}
	}

	s.Halt(context) // All children succeeded, reset before returning Success
	return Success
}

// Halt resets the running child index for the SequenceNode and halts its children.
func (s *SequenceNode) Halt(context *TickContext) {
	s.runningChildIndex = 0
	s.CompositeNode.Halt(context) // Call base Halt to halt all children
}


// --- Selector Node ---

// SelectorNode ticks its children sequentially until one of them returns Success.
// If any child returns Success, the SelectorNode immediately returns Success.
// If any child returns Running, the SelectorNode immediately returns Running, remembering the running child.
// If all children return Failure, the SelectorNode returns Failure.
type SelectorNode struct {
	CompositeNode
	runningChildIndex int // Index of the child that is currently running
}

// NewSelector creates a new SelectorNode.
func NewSelector(children ...Node) *SelectorNode {
	s := &SelectorNode{runningChildIndex: 0}
	s.Type = "Selector"
	s.AddChildren(children...)
	return s
}

// Tick executes the children of the SelectorNode.
func (s *SelectorNode) Tick(context *TickContext) Status {
	for i := s.runningChildIndex; i < len(s.Children); i++ {
		child := s.Children[i]
		status := child.Tick(context)

		switch status {
		case Success:
			s.Halt(context) // Halt all children and reset self before returning
			return Success
		case Running:
			s.runningChildIndex = i // Remember the running child
			return Running
		case Failure:
			// Continue to the next child
		default: // Invalid status
			s.Halt(context)
			return Failure // Or some error status if defined
		}
	}
	s.Halt(context) // All children failed, reset before returning Failure
	return Failure
}

// Halt resets the running child index for the SelectorNode and halts its children.
func (s *SelectorNode) Halt(context *TickContext) {
	s.runningChildIndex = 0
	s.CompositeNode.Halt(context) // Call base Halt to halt all children
}
