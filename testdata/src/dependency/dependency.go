package dependency // want "dependent analyzer ran"

// This file is used to test analyzer dependencies
func TestFunc() { // want "TestFunc is public"
	// This should trigger both analyzers
}

var PublicVar = 42 // want "PublicVar is public"
