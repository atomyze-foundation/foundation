package acl

// Role is the role of the user
type Role string

// String returns the string representation of the role
func (r Role) String() string {
	return string(r)
}

const (
	// Issuer is the issuer role
	Issuer Role = "issuer"
	// FeeSetter is the fee setter role
	FeeSetter Role = "feeSetter"
)
