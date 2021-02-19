package cluster

import (
	"fmt"

	"github.com/zalando/postgres-operator/pkg/spec"
	v1 "k8s.io/api/core/v1"
)

// VersionMap Map of version numbers
var VersionMap = map[string]int{
	"9.5": 9500,
	"9.6": 9600,
	"10":  10000,
	"11":  11000,
	"12":  12000,
	"13":  13000,
}

// IsBiggerPostgresVersion Compare two Postgres version numbers
func IsBiggerPostgresVersion(old string, new string) bool {
	oldN, _ := VersionMap[old]
	newN, _ := VersionMap[new]
	return newN > oldN
}

// GetDesiredMajorVersionAsInt Convert string to comparable integer of PG version
func (c *Cluster) GetDesiredMajorVersionAsInt() int {
	return VersionMap[c.GetDesiredMajorVersion()]
}

// GetDesiredMajorVersion returns major version to use, incl. potential auto upgrade
func (c *Cluster) GetDesiredMajorVersion() string {

	if c.Config.OpConfig.MajorVersionUpgradeMode == "full" {
		if IsBiggerPostgresVersion(c.Spec.PgVersion, c.Config.OpConfig.TargetMajorVersion) {
			c.logger.Infof("Overwriting configured major version %s to %s", c.Spec.PgVersion, c.Config.OpConfig.TargetMajorVersion)
			return c.Config.OpConfig.TargetMajorVersion
		}
	}

	return c.Spec.PgVersion
}

func (c *Cluster) majorVersionUpgrade() error {

	if c.OpConfig.MajorVersionUpgradeMode == "off" {
		return nil
	}

	pods, _ := c.listPods()
	allRunning := true

	var masterPod *v1.Pod

	for _, pod := range pods {
		ps, _ := c.patroni.GetMemberData(&pod)

		if ps.State != "running" {
			allRunning = false
		}

		if ps.Role == "master" {
			masterPod = &pod
			c.currentMajorVersion = ps.ServerVersion
		}
	}

	numberOfPods := len(pods)
	if allRunning && masterPod != nil {
		if c.currentMajorVersion < c.GetDesiredMajorVersionAsInt() {
			podName := &spec.NamespacedName{Namespace: masterPod.Namespace, Name: masterPod.Name}
			c.ExecCommand(podName, fmt.Sprintf("python3 /scripts/inplace_upgrade.py %d", numberOfPods))
		}
	}

	return nil
}

func (c *Cluster) getCurrentMajorVersion() error {
	return nil
}