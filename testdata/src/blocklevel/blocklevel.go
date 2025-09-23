package blocklevel

func goodFunc() {
	var normalVar int
	_ = normalVar
}

//ignore:shortnames
func suppressedFunc() {
	var x int // should be suppressed
	_ = x
}

func unsuppressedFunc() {
	var y int // want "y is single letter"
	_ = y
}

//ignore:underscorenames
func nestedSuppressedFunc() {
	func_with_underscores := func() { // should be suppressed
		var snake_case int // should be suppressed
		_ = snake_case
	}
	func_with_underscores()
}

func nestedUnsuppressedFunc() {
	func_with_underscores := func() { // want "func_with_underscores has underscore"
		var snake_case int // want "snake_case has underscore"
		_ = snake_case
	}
	func_with_underscores()
}

//ignore:all
func allSuppressedFunc() {
	var z int // should be suppressed
	_ = z

	another_func := func() { // should be suppressed
		var under_score int // should be suppressed
		_ = under_score
	}
	another_func()
}
