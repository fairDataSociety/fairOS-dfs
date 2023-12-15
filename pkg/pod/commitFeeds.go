package pod

// CommitFeeds commits uncommitted feeds saved in local lru cache to swarm
func (p *Pod) CommitFeeds(podName string) error {
	podInfo, _, err := p.GetPodInfoFromPodMap(podName)
	if err != nil { // skipcq: TCV-001
		return err
	}
	podInfo.feed.CommitFeeds()
	p.fd.CommitFeeds()
	return nil
}
