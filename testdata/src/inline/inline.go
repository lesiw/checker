package inline

func goodFunc() {}

var normalVar int //ignore:numberednames

func PublicFunc() {} //ignore:publicnames
var count2 int    //ignore:numberednames
var count3 int    // want "count3 has numbers"

var anotherVar int //ignore:all

func AnotherPublic() {} //ignore:all
var value3 int       //ignore:all

var yetAnotherVar int //ignore:publicnames,numberednames

func YetAnotherPublic() {} //ignore:publicnames,numberednames
var item4 string        //ignore:publicnames,numberednames

var commentedVar int //ignore:numberednames // This is legacy code

func CommentedPublic() {} //ignore:publicnames // Complex legacy function

func _() {
	if count5 := count2; count5 != count2 { //ignore:numberednames
	}
	eq := func(x, y int) bool { return x == y }
	if count6 := count2; !eq(count6, count2) { //ignore:numberednames
	}
}
