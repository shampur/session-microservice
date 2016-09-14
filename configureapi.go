package session

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type apiConfig struct {
	routelist	[]routedetail
}

type routedetail struct {
	Api string		`json:"api"`
	Methods []string	`json:"methods"`
	Destination string	`json:"destination"`
	Authorization bool	`json:"authorization"`
}

func GetApiConfig() *apiConfig {
	return &apiConfig{
		routelist: getapidetails(Apiconfigfile),
	}
}


func getapidetails(configfile string) []routedetail {
	file, e := ioutil.ReadFile(configfile)
	if e != nil {
		fmt.Println("Error in reading file")
		return []routedetail{}
	}

	var jsondata []routedetail
	json.Unmarshal(file, &jsondata)
	return jsondata
}