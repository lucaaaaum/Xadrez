package tcp

import "context"

type peerKey struct{}

func peerFrom(ctx context.Context) *Peer {
	peer, _ := ctx.Value(peerKey{}).(*Peer)
	return peer
}

func SetUser(ctx context.Context, user string) {
	peer := peerFrom(ctx)
	if peer == nil {
		return
	}
	peer.mu.Lock()
	defer peer.mu.Unlock()
	peer.user = user
}

func UserFrom(ctx context.Context) (string, bool) {
	peer := peerFrom(ctx)
	if peer == nil {
		return "", false
	}
	peer.mu.Lock()
	defer peer.mu.Unlock()
	return peer.user, peer.user != ""
}
