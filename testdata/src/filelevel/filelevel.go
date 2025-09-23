//ignore:publicnames
package filelevel

func goodFunc() {}

func PublicFunc() {} // should be suppressed by file-level directive

var count2 int // want "count2 has numbers"
