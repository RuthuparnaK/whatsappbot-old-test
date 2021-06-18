package config

type Databaseconfig struct {
	username, password, host, port string
}

type Severconfig struct {
	port string
}

func (b *Databaseconfig) Assigndb() (username string, password string, host string, port string) {
	// staging
	b.username = "www-data"
	b.password = "krishna_giri"
	b.host = "localhost"
	b.port = "5432"

	//development - pavan
	// b.username = "postgres"
	// b.password = "postgres"
	// b.host = "localhost"
	// b.port = "5432"

	//development - ruthu
	// b.username = "latlong"
	// b.password = "latlong123"
	// b.host = "localhost"
	// b.port = "5432"

	return b.username, b.password, b.host, b.port
}

func (b *Severconfig) Assignserver() (port string) {
	// staging
	// b.port = "5050"

	// development
	b.port = "3000"
	return b.port
}
