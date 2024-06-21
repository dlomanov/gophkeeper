package md

const (
	Schema  = "bearer"
	AuthKey = "authorization"
)

func NewTokenKV(token string) []string {
	return []string{AuthKey, Schema + " " + token}
}
