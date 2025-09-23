package block

func goodFunc() {}

//ignore:publicnames
func PublicFunc() {}

//ignore:numberednames
var count2 int

//ignore:all
func AllPublicFunc() {}

//ignore:all
var value3 int

//ignore:publicnames,numberednames
func MultiplePublic() {}

//ignore:publicnames,numberednames
var item4 string

//ignore:publicnames // This legacy function is complex
func CommentedPublic() {}

//ignore:all
var (
	badVar1 int
	badVar2 string
)

//ignore:numberednames
var (
	specificVar1 int
	specificVar2 string
)
