package inline

func goodFunc() {}

var normalVar int //ignore:numberednames

func PublicFunc() {} //ignore:publicnames
var count2 int    //ignore:numberednames

var anotherVar int //ignore:all

func AnotherPublic() {} //ignore:all
var value3 int       //ignore:all

var yetAnotherVar int //ignore:publicnames,numberednames

func YetAnotherPublic() {} //ignore:publicnames,numberednames
var item4 string        //ignore:publicnames,numberednames

var commentedVar int //ignore:numberednames // This is legacy code

func CommentedPublic() {} //ignore:publicnames // Complex legacy function
