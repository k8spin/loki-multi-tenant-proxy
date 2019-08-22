package pkg

import (
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Authn Contains a list of users
type Authn struct {
	Users []User `yaml:"users"`
}

// User Identifies an username
type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	OrgID    string `yaml:"orgid"`
}

// ParseConfig read a configuration file in the path `location` and returns an Authn object
func ParseConfig(location *string) (*Authn, error) {
	data, err := ioutil.ReadFile(*location)
	if err != nil {
		return nil, err
	}
	authn := Authn{}
	err = yaml.Unmarshal([]byte(data), &authn)
	if err != nil {
		return nil, err
	}
	return &authn, nil
}

// GetOrgID Returns the org id from a given username
func GetOrgID(userName string, users *Authn) (string, error) {
	for _, v := range users.Users {
		if v.Username == userName {
			return v.OrgID, nil
		}
	}
	return "", errors.New("User not found")
}
