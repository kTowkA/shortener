package pkg03

func xrun() {
	yrun()
}
func yrun() int {
	zrun() // want "not recommended function"
	return 0
}
func zrun() int {
	return 0
}
