/*
 * UDR Configuration Factory
 */

package factory

type Config struct {
	Info *Info `yaml:"info"`

	Configuration *Configuration `yaml:"configuration"`
}

type Info struct {
	Version string `yaml:"version,omitempty"`

	Description string `yaml:"description,omitempty"`
}

type Configuration struct {
	Sbi *Sbi `yaml:"sbi"`

	Mongodb *Mongodb `yaml:"mongodb"`

	NrfUri string `yaml:"nrfUri"`
}

type Sbi struct {
	Scheme   string `yaml:"scheme"`
	IPv4Addr string `yaml:"ipv4Addr"`
	// IPv6Addr string `yaml:"ipv6Addr,omitempty"`
	Port int  `yaml:"port"`
	Tls  *Tls `yaml:"tls,omitempty"`
}

type Tls struct {
	Log string `yaml:"log"`

	Pem string `yaml:"pem"`

	Key string `yaml:"key"`
}

type Mongodb struct {
	Name string `yaml:"name"`

	Url string `yaml:"url"`
}
