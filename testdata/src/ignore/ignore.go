package ignore

func goodFunc() {}

//ignore:publicnames
func PublicFunc() {} // should be suppressed

//ignore:numberednames
var count1 int // should be suppressed

func AnotherPublic() {} // want "publicnames: AnotherPublic is public"

var item2 string // want "numberednames: item2 has numbers"
