package linestart

func reportedFunc() {
	var x int
	_ = x

	// TODO reported // want "comment contains TODO"
}

func suppressedFunc() {
	// TODO the diagnostic anchors at column 1, left of this
	// indented comment group.
	//ignore:todocomments
	var x int
	_ = x
}
