package acl

type Role string

func (r Role) String() string {
	return string(r)
}

const (
	Issuer    Role = "issuer"
	FeeSetter Role = "feeSetter"
)
