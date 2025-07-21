package auth

// APIUser represents user data returned from the API
type APIUser struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Avatar    string
}
