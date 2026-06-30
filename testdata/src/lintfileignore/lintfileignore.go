package lintfileignore

//lint:file-ignore publicnames generated bindings, names are external

func goodFunc() {}

func PublicFunc() {} // should be suppressed by file-level directive

var count2 int // want "count2 has numbers"
