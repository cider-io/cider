package gossip

func GetMembershipList() map[string]Member {
	return Self.MembershipList
}
