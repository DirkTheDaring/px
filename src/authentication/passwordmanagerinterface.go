package authentication

type PasswordManager interface {
	GetCredentials(urlString, domain string, username string) (string, error)
}
