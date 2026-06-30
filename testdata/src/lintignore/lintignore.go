package lintignore

func goodFunc() {}

//lint:ignore publicnames legacy public name we cannot rename
func PublicFunc() {} // should be suppressed

//lint:ignore numberednames numeric suffix is part of an external contract
var count1 int // should be suppressed

//lint:ignore publicnames,numberednames both are legacy
func MultiplePublic() {} // should be suppressed

var inlineVar1 int //lint:ignore numberednames numeric for sort order

func InlinePublic() {} //lint:ignore publicnames API stability

func AnotherPublic() {} // want "AnotherPublic is public"

var item2 string // want "item2 has numbers"
