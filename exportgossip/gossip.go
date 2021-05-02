package exportgossip

import (
	"cider/gossip"
)

func GetMembershipList() map[string]gossip.Member {
	return gossip.Self.MembershipList
}
