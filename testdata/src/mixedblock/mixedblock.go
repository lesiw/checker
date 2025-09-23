package mixedblock

func goodFunc() {
	var normalVar int
	_ = normalVar
}

// Function declaration should trigger publicnames, but content should be suppressed
//
//ignore:shortnames,underscorenames
func PublicFunc() { // want "PublicFunc is public"
	var x int // should be suppressed by block ignore
	_ = x

	func_with_underscores := func() { // should be suppressed by block ignore
		var snake_case int // should be suppressed by block ignore
		_ = snake_case
	}
	func_with_underscores()
}

// Function declaration should NOT trigger publicnames, but content should trigger shortnames/underscorenames
func normalFunc() {
	var y int // want "y is single letter"
	_ = y

	func_with_underscores := func() { // want "func_with_underscores has underscore"
		var snake_case int // want "snake_case has underscore"
		_ = snake_case
	}
	func_with_underscores()
}

// Everything should be suppressed
//
//ignore:all
func AllSuppressedFunc() {
	var z int // should be suppressed
	_ = z
	var count1 int // should be suppressed
	_ = count1

	another_func := func() { // should be suppressed
		var under_score int // should be suppressed
		_ = under_score
	}
	another_func()
}

// Only specific analyzers suppressed
//
//ignore:publicnames,shortnames
func PartiallySupressed() {
	var a int // should be suppressed by shortnames ignore
	_ = a
	var count2 int // want "count2 has numbers"
	_ = count2

	func_with_underscores := func() { // want "func_with_underscores has underscore"
		var snake_case int // want "snake_case has underscore"
		_ = snake_case
	}
	func_with_underscores()
}
