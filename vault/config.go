package vault

type Share struct {
	Name string
	Dir  string
}

type Config struct {
	Shares []Share
}

func (c *Config) AddShare(name, dir string) {
	c.Shares = append(c.Shares, Share{Name: name, Dir: dir})
}
