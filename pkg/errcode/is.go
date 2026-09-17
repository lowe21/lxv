package errcode

func Is(err error, target error) bool {
	if err == nil || target == nil {
		return false
	}

	errSubCode, _ := Parse(New(err))
	targetSubCode, _ := Parse(New(target))

	return errSubCode == targetSubCode
}
