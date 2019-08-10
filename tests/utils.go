package tests

func RunSilent(f func(id string) error, v string) {
	_ = f(v)
}
