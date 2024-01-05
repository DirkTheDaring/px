package authentication

type GetCredentialsCallback func(urlString string, domain string, username string) (string, error)

type PasswordManager interface {
	GetCredentials(urlString, domain string, username string) (string, error)
}
