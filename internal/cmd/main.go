package cmd

func Main() {
	c, err := NewConfig()
	if err != nil {
		panic(err.Error())
	}

	startNotifier(c)
}
