package nolint

func goodFunc() {}

//nolint:publicnames
func PublicFunc() {} // should be suppressed

//nolint:numberednames
var count1 int // should be suppressed

//nolint:all
func AllPublicFunc() {} // should be suppressed

//nolint:all
var value3 int // should be suppressed

//nolint:publicnames,numberednames
func MultiplePublic() {} // should be suppressed

//nolint:publicnames,numberednames
var item4 string // should be suppressed

//nolint:publicnames // legacy: complex function
func CommentedPublic() {} // should be suppressed

var inlineVar1 int //nolint:numberednames

func InlinePublic() {} //nolint:publicnames

func AnotherPublic() {} // want "AnotherPublic is public"

var item2 string // want "item2 has numbers"
