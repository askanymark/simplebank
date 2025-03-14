package util

const (
	DepositorRole = "depositor"
	BankerRole    = "banker"
)

func IsBanker(role string) bool {
	return role == BankerRole
}

func IsDepositor(role string) bool {
	return role == DepositorRole
}
