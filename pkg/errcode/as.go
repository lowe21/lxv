package errcode

func As(err error, target error) bool {
	errSubCode, _ := Parse(New(err))
	targetSubCode, _ := Parse(New(target))

	return errSubCode != "" && errSubCode == targetSubCode
}
