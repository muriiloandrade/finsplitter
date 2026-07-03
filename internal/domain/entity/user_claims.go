package entity

// UserClaims holds the JWT claims extracted from the Logto token.
// Name, phone, and picture are populated via Logto's Custom JWT script.
type UserClaims struct {
	Sub      string `json:"sub"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Picture  string `json:"picture,omitempty"`
}
