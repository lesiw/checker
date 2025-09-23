package ignore

func goodFunc() {}

//ignore:publicnames
func PublicFunc() {} // should be suppressed

//ignore:numberednames
var count1 int // should be suppressed

func AnotherPublic() {} // want "AnotherPublic is public"

var item2 string // want "item2 has numbers"
