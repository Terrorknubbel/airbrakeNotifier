package cmd

func Main() {
	c, err := newConfig()
	if err != nil {
		panic(err.Error())
	}

	startNotifier(c)
}
