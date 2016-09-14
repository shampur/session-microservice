package session

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type localAuth struct {
	localAuthFileData 	[]fileFormat
}

type fileFormat struct {
	Username	string 	`json:"username"`
	Password	string	`json:"password"`
	Active		bool	`json:"active"`

}

// NewLdap function initializes the ldap module
func NewLocalAuth() *localAuth {
	return &localAuth{
		localAuthFileData: 	getSuperAdminUsers(LocalAuthFileLoc),
	}
}

func getSuperAdminUsers(filepath string) []fileFormat {
	file, e := ioutil.ReadFile(filepath)
    	if e != nil {
		fmt.Println("Error in reading file")
        	return []fileFormat{}
    	}
    	var jsondata []fileFormat
    	json.Unmarshal(file, &jsondata)
    	return jsondata
}

func (l *localAuth) authenticate(cred Credentials) (LoginResponse, error) {
	for _, element := range l.localAuthFileData {
		if (element.Username == cred.Username){
			if(element.Password == cred.Password && element.Active==true){
				return LoginResponse{Authenticated: true, Message: "success"}, nil
			}
		}
	}
	return LoginResponse{Authenticated: false, Message: "Invalid username or password"}, nil
}
