package session


//AuthManager is responsible for cycling through the different authentication mechanizms
type AuthManager struct {
	authModuleCount int64
	authModules     []AuthInterface
}

func (a *AuthManager) authenticate(cred Credentials) (LoginResponse, error) {
	var result LoginResponse
	var err error

	for _, element := range a.authModules {
		result, err = element.authenticate(cred)
		if result.Authenticated{
			return result, err
		}
	}
	return LoginResponse{Authenticated: false, Message: "Invalid username or password"}, nil
}

//NewAuthmanager creates a new authentication manager
func NewAuthmanager() *AuthManager {
	return &AuthManager{
		authModuleCount: 1,
		authModules:     createModules(),
	}
}

func createModules() []AuthInterface {
	var inter []AuthInterface
	ldapModule := NewLdapAuth()
	inter = append(inter, ldapModule)
	localAuthModule := NewLocalAuth();
	inter = append(inter, localAuthModule)
	return inter
}

//AuthInterface - all authentication modules should implement this interface
type AuthInterface interface {
	authenticate(cred Credentials) (LoginResponse, error)
}

