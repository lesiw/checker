package commentignore

// TODO: This should be reported // want "comment contains TODO"
func goodFunc() {}

// TODO: This should be suppressed //ignore:todocomments
func inlineSuppressedFunc() {}

// TODO: This should be suppressed
//
//ignore:todocomments
func blockSuppressedFunc() {}

// TODO: This should be suppressed
//
//ignore:all
func allSuppressedFunc() {}

// TODO: This should be reported // want "comment contains TODO"
func normalFunc() {}

// TODO: This should be suppressed
// TODO: This should also be suppressed
//
//ignore:todocomments
func multiLineSuppressedFunc() {}

// TODO: Multiple comments should be reported // want "comment contains TODO"
// TODO: This should be reported // want "comment contains TODO"
// TODO: This should also be reported // want "comment contains TODO"
func multiLineNormalFunc() {}

// TODO: This is a standalone comment block with no associated code
// TODO: This comment should be suppressed by ignore
//
//ignore:todocomments

func separatorFunc() {}

// TODO: This is another standalone comment that should be reported // want "comment contains TODO"

//ignore:todocomments
func inheritanceTestFunc() {
	// TODO: This comment inside the function should inherit the ignore
	// TODO: This comment should also be suppressed by inheritance
	var x int
	_ = x
}
