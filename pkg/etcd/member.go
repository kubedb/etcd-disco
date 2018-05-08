package etcd

/*import (
	"context"

	etcdcl "github.com/coreos/etcd/clientv3"
	"github.com/etcd-manager/lector/pkg/cmds/options"
	"github.com/etcd-manager/lector/pkg/util"
	log "github.com/sirupsen/logrus"
)

func JoinMember(opts *options.EtcdServerJoinConfig) error {

*/ /*	conf := ServerConfig{
	Name: opts.Name,
	DataDir: fmt.Sprintf("/var/lib/etcd/%v", opts.Name),
	DataQuota: 2147483648,
	Pe


}*/ /*

	client, err := etcdcl.New(etcdcl.Config{
		Endpoints:   opts.InitialUrls,
		DialTimeout: util.DefaultDialTimeout,
		// TLS:              tc,
		AutoSyncInterval: util.DefaultAutoSync,
	})
	if err != nil {
		return err
	}

	// Set the internal configuration.
	initialPURLs := map[string]string{
		opts.Name: opts.PeerUrl,
	}

	ctx, cancel := context.WithTimeout(context.Background(), util.DefaultStartRejoinTimeout)
	defer cancel()

	members, err := client.MemberList(ctx)
	if err != nil {
		return err
	}

	for _, member := range members.Members {
		if member.Name == "" {
			continue
		}
		initialPURLs[member.Name] = member.PeerURLs[0]
	}

	// Check if we are listed as a member, and save the member ID if so.
	var memberID uint64
	for _, member := range members.Members {
		if opts.Name == member.Name {
			memberID = member.ID
			break
		}
	}

	//snapshot add

	if memberID != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), defaultStartRejoinTimeout)
		defer cancel()
*/ /*if err := c.startServer(ctx); err == nil {
	return nil
}*/ /*

		//log.Warn("failed to join as an existing member, resetting")
		if err := client.RemoveMember(c.cfg.Name, memberID); err != nil {
			log.WithError(err).Warning("failed to remove ourselves from the cluster's member list")
		}
	}

}*/
