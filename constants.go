package session

import "errors"

var (
	// ErrInconsistentIDs server error message
	ErrInconsistentIDs = errors.New("inconsistent IDs")
	// ErrAlreadyExists server error message
	ErrAlreadyExists = errors.New("already exists")
	// ErrNotFound server error message
	ErrNotFound = errors.New("not found")
)

var (
	//Local authorization file location
	LocalAuthFileLoc = "localauthfile.json"
	//api configuration file
	Apiconfigfile = "apiconfig.json"
	// Session timeout
	SessionTimeOut = 0.4
)
