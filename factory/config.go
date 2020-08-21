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
	Scheme       string `yaml:"scheme"`
	RegisterIPv4 string `yaml:"registerIPv4,omitempty"` // IP that is registered at NRF.
	// IPv6Addr string `yaml:"ipv6Addr,omitempty"`
	BindingIPv4 string `yaml:"bindingIPv4,omitempty"` // IP used to run the server in the node.
	Port        int    `yaml:"port"`
	Tls         *Tls   `yaml:"tls,omitempty"`
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
