package multiple

func goodFunc() {}

//ignore:publicnames,numberednames
func PublicFunc() {} // should be suppressed

//ignore:publicnames,numberednames
var count1 int // should be suppressed

func AnotherPublic() {} // want "AnotherPublic is public"

var item2 string // want "item2 has numbers"

//ignore:all
func ThirdPublic() {} // should be suppressed

//ignore:all
var value3 int // should be suppressed
