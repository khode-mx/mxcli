package tui

// ViewStack manages a stack of Views on top of a base view.
// The base view is always present; pushed views overlay it.
type ViewStack struct {
	base  View
	stack []View
}

// NewViewStack creates a ViewStack with the given base view.
func NewViewStack(base View) ViewStack {
	return ViewStack{base: base}
}

// Active returns the top of the stack, or the base view if the stack is empty.
func (vs *ViewStack) Active() View {
	if len(vs.stack) > 0 {
		return vs.stack[len(vs.stack)-1]
	}
	return vs.base
}

// Push adds a view to the top of the stack.
func (vs *ViewStack) Push(v View) {
	vs.stack = append(vs.stack, v)
}

// Pop removes and returns the top view from the stack.
// Returns false if the stack is empty (base view is never popped).
func (vs *ViewStack) Pop() (View, bool) {
	if len(vs.stack) == 0 {
		return nil, false
	}
	top := vs.stack[len(vs.stack)-1]
	vs.stack = vs.stack[:len(vs.stack)-1]
	return top, true
}

// Depth returns the total number of views (base + stacked).
func (vs *ViewStack) Depth() int {
	return len(vs.stack) + 1
}

// SetActive replaces the top of the stack, or the base view if the stack is empty.
func (vs *ViewStack) SetActive(v View) {
	if len(vs.stack) > 0 {
		vs.stack[len(vs.stack)-1] = v
	} else {
		vs.base = v
	}
}
