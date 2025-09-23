package basic

func goodFunc() {}

func PublicFunc() {} // want "PublicFunc is public"

var count1 int // want "count1 has numbers"
var normalVar int

func AnotherPublic() {} // want "AnotherPublic is public"

var item2 string // want "item2 has numbers"
