package basic

func goodFunc() {}

func PublicFunc() {} // want "publicnames: PublicFunc is public"

var count1 int // want "numberednames: count1 has numbers"
var normalVar int

func AnotherPublic() {} // want "publicnames: AnotherPublic is public"

var item2 string // want "numberednames: item2 has numbers"
