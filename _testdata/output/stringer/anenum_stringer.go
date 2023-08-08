package stringer

var _AnEnum_strings = []string{
	"Val1",
	"Val2",
	"Val3",
	}

func (self AnEnum) String() string { return _AnEnum_strings[int(self)] }
