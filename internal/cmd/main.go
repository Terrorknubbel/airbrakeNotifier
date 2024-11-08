package cmd

func Main() {
	c, err := newConfig()
	if err != nil {
		c.logger.Error(err)
		return
	}

	startNotifier(c)
}
