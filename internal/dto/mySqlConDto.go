package dto

type MySQLConnectionDto struct {
	Dns                   string
	ConnectionMaxLifetime int
	MaxOpenConnections    int
	MaxIdleConnections    int
}

func NewMySQLConnectionDto(dns string, maxLft int, maxOptConns int, maxIdleConns int) *MySQLConnectionDto {
	return &MySQLConnectionDto{
		Dns:                   dns,
		ConnectionMaxLifetime: maxLft,
		MaxOpenConnections:    maxOptConns,
		MaxIdleConnections:    maxIdleConns,
	}
}
