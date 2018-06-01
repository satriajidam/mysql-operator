// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"github.com/oracle/mysql-operator/pkg/constants"
	"github.com/oracle/mysql-operator/pkg/version"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// The default MySQL version to use if not specified explicitly by user
	defaultVersion      = "8.0.11"
	defaultReplicas     = 1
	defaultBaseServerID = 1000
	// Max safe value for BaseServerID calculated as max MySQL server_id value - max Replication Group size
	maxBaseServerID uint32 = 4294967295 - 9
)

const (
	// MaxInnoDBClusterMembers is the maximum number of members supported by InnoDB
	// group replication.
	MaxInnoDBClusterMembers = 9

	// ClusterNameMaxLen is the maximum supported length of a
	// Cluster name.
	// See: https://bugs.mysql.com/bug.php?id=90601
	ClusterNameMaxLen = 28
)

// TODO (owain) we need to remove this because it's not reasonable for us to maintain a list
// of all the potential MySQL versions that can be used and in reality, it shouldn't matter
// too much. The burden of this is not worth the benfit to a user
var validVersions = []string{
	defaultVersion,
}

// ClusterSpec defines the attributes a user can specify when creating a cluster
type ClusterSpec struct {
	// Version defines the MySQL Docker image version.
	Version string `json:"version"`

	// Replicas defines the number of running MySQL instances in a cluster
	Replicas int32 `json:"replicas,omitempty"`

	// BaseServerID defines the base number used to create uniq server_id for MySQL instances in a cluster.
	// The baseServerId value need to be in range from 1 to 4294967286
	// If ommited in the manifest file, or set to 0, defaultBaseServerID value will be used.
	BaseServerID uint32 `json:"baseServerId,omitempty"`

	// MultiMaster defines the mode of the MySQL cluster. If set to true,
	// all instances will be R/W. If false (the default), only a single instance
	// will be R/W and the rest will be R/O.
	MultiMaster bool `json:"multiMaster,omitempty"`

	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// If specified, affinity will define the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// VolumeClaimTemplate allows a user to specify how volumes inside a MySQL cluster
	// +optional
	VolumeClaimTemplate *corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`

	// BackupVolumeClaimTemplate allows a user to specify a volume to temporarily store the
	// data for a backup prior to it being shipped to object storage.
	// +optional
	BackupVolumeClaimTemplate *corev1.PersistentVolumeClaim `json:"backupVolumeClaimTemplate,omitempty"`

	// If defined, we use this secret for configuring the MYSQL_ROOT_PASSWORD
	// If it is not set we generate a secret dynamically
	// +optional
	SecretRef *corev1.LocalObjectReference `json:"secretRef,omitempty"`

	// ConfigRef allows a user to specify a custom configuration file for MySQL.
	// +optional
	ConfigRef *corev1.LocalObjectReference `json:"configRef,omitempty"`

	// SSLSecretRef allows a user to specify custom CA certificate, server certificate
	// and server key for group replication SSL
	// +optional
	SSLSecretRef *corev1.LocalObjectReference `json:"sslSecretRef,omitempty"`
}

// ClusterPhase describes the state of the cluster.
type ClusterPhase string

const (
	// ClusterPhasePending means the cluster has been accepted by the system,
	// but one or more of the services or statefulsets has not been started.
	// This includes time before being bound to a node, as well as time spent
	// pulling images onto the host.
	ClusterPhasePending ClusterPhase = "Pending"

	// ClusterPhaseRunning means the cluster has been created, all of it's
	// required components are present, and there is at least one endpoint that
	// mysql client can connect to.
	ClusterPhaseRunning ClusterPhase = "Running"

	// ClusterPhaseSucceeded means that all containers in the pod have
	// voluntarily terminated with a container exit code of 0, and the system
	// is not going to restart any of these containers.
	ClusterPhaseSucceeded ClusterPhase = "Succeeded"

	// ClusterPhaseFailed means that all containers in the pod have terminated,
	// and at least one container has terminated in a failure (exited with a
	// non-zero exit code or was stopped by the system).
	ClusterPhaseFailed ClusterPhase = "Failed"

	// ClusterPhaseUnknown means that for some reason the state of the cluster
	// could not be obtained, typically due to an error in communicating with
	// the host of the pod.
	ClusterPhaseUnknown ClusterPhase = ""
)

// ValidClusterPhases denote the life-cycle states a cluster can be in.
var ValidClusterPhases = []ClusterPhase{
	ClusterPhasePending,
	ClusterPhaseRunning,
	ClusterPhaseSucceeded,
	ClusterPhaseFailed,
	ClusterPhaseUnknown,
}

// ClusterStatus defines the current status of a MySQL cluster
// propagating useful information back to the cluster admin
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Phase             ClusterPhase `json:"phase"`
	Errors            []string     `json:"errors"`
}

// +genclient
// +genclient:noStatus

// Cluster represents a cluster spec and associated metadata
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ClusterSpec   `json:"spec"`
	Status            ClusterStatus `json:"status"`
}

// ClusterList is a placeholder type for a list of MySQL clusters
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Cluster `json:"items"`
}

// Validate returns an error if a cluster is invalid
func (c *Cluster) Validate() error {
	return validateCluster(c).ToAggregate()
}

// EnsureDefaults will ensure that if a user omits and fields in the
// spec that are required, we set some sensible defaults.
// For example a user can choose to omit the version
// and number of replics
func (c *Cluster) EnsureDefaults() *Cluster {
	if c.Spec.Replicas == 0 {
		c.Spec.Replicas = defaultReplicas
	}

	if c.Spec.BaseServerID == 0 {
		c.Spec.BaseServerID = defaultBaseServerID
	}

	if c.Spec.Version == "" {
		c.Spec.Version = defaultVersion
	}

	return c
}

// RequiresConfigMount will return true if a user has specified a config map
// for configuring the cluster else false
func (c *Cluster) RequiresConfigMount() bool {
	return c.Spec.ConfigRef != nil
}

// RequiresSecret returns true if a secret should be generated
// for a MySQL cluster else false
func (c *Cluster) RequiresSecret() bool {
	return c.Spec.SecretRef == nil
}

// RequiresCustomSSLSetup returns true is the user has provided a secret
// that contains CA cert, server cert and server key for group replication
// SSL support
func (c *Cluster) RequiresCustomSSLSetup() bool {
	return c.Spec.SSLSecretRef != nil
}

// BackupSpec defines the specification for a MySQL backup. This includes what should be backed up,
// what tool should perform the backup, and, where the backup should be stored.
type BackupSpec struct {
	// Executor is the configuration of the tool that will produce the backup, and a definition of
	// what databases and tables to backup.
	Executor *BackupExecutor `json:"executor"`

	// StorageProvider is the configuration of where and how backups should be stored.
	StorageProvider *BackupStorageProvider `json:"storageProvider"`

	// Cluster is a reference to the Cluster to which the Backup belongs.
	Cluster *corev1.LocalObjectReference `json:"cluster"`

	// AgentScheduled is the agent hostname to run the backup on.
	// TODO(apryde): ScheduledAgent (*corev1.LocalObjectReference)?
	AgentScheduled string `json:"agentscheduled"`
}

// BackupExecutor represents the configuration of the tool performing the backup. This includes the tool
// to use, and, what database and tables should be backed up.
// The storage of the backup is configured in the relevant Storage configuration.
type BackupExecutor struct {
	// Name of the tool performing the backup, e.g. mysqldump.
	Name string `json:"name"`
	// Databases are the databases to backup.
	Databases []string `json:"databases"`
}

// BackupStorageProvider defines the configuration for storing a MySQL backup to a storage service.
// The generation of the backup is configured in the Executor configuration.
type BackupStorageProvider struct {
	// Name denotes the type of storage provider that will store and retrieve the backups,
	// e.g. s3, oci-s3-compat, aws-s3, gce-s3, etc.
	Name string `json:"name"`
	// SecretRef is a reference to the Kubernetes secret containing the configuration for uploading
	// the backup to authenticated storage.
	SecretRef *corev1.LocalObjectReference `json:"secretRef,omitempty"`
	// Config is generic string based key-value map that defines non-secret configuration values for
	// uploading the backup to storage w.r.t the configured storage provider.
	Config map[string]string `json:"config,omitempty"`
}

// BackupPhase represents the current life-cycle phase of a Backup.
type BackupPhase string

const (
	// BackupPhaseUnknown means that the backup hasn't yet been processed.
	BackupPhaseUnknown BackupPhase = ""

	// BackupPhaseNew means that the Backup hasn't yet been processed.
	BackupPhaseNew BackupPhase = "New"

	// BackupPhaseScheduled means that the Backup has been scheduled on an
	// appropriate replica.
	BackupPhaseScheduled BackupPhase = "Scheduled"

	// BackupPhaseStarted means the backup is in progress.
	BackupPhaseStarted BackupPhase = "Started"

	// BackupPhaseComplete means the backup has terminated successfully.
	BackupPhaseComplete BackupPhase = "Complete"

	// BackupPhaseFailed means the backup has terminated with an error.
	BackupPhaseFailed BackupPhase = "Failed"
)

// BackupOutcome describes the location of a MySQL Backup
type BackupOutcome struct {
	// Location is the Object Storage network location of the Backup.
	Location string `json:"location"`
}

// BackupStatus captures the current status of a MySQL backup.
type BackupStatus struct {
	// Phase is the current life-cycle phase of the Backup.
	Phase BackupPhase `json:"phase"`

	// Outcome holds the results of a successful backup.
	Outcome BackupOutcome `json:"outcome"`

	// TimeStarted is the time at which the backup was started.
	TimeStarted metav1.Time `json:"timeStarted"`

	// TimeCompleted is the time at which the backup completed.
	TimeCompleted metav1.Time `json:"timeCompleted"`
}

// +genclient
// +genclient:noStatus

// Backup is a MySQL Operator resource that represents a backup of a MySQL
// cluster.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Backup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   BackupSpec   `json:"spec"`
	Status BackupStatus `json:"status"`
}

// BackupList is a list of Backups.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Backup `json:"items"`
}

// EnsureDefaults can be invoked to ensure the default values are present.
func (b Backup) EnsureDefaults() *Backup {
	buildVersion := version.GetBuildVersion()
	if buildVersion != "" {
		if b.Labels == nil {
			b.Labels = make(map[string]string)
		}
		_, hasKey := b.Labels[constants.MySQLOperatorVersionLabel]
		if !hasKey {
			SetOperatorVersionLabel(b.Labels, buildVersion)
		}
	}
	return &b
}

// Validate checks if the resource spec is valid.
func (b Backup) Validate() error {
	return validateBackup(&b).ToAggregate()
}

// BackupScheduleSpec defines the specification for a MySQL backup schedule.
type BackupScheduleSpec struct {
	// Schedule specifies the cron string used for backup scheduling.
	Schedule string `json:"schedule"`

	// BackupTemplate is the specification of the backup structure
	// to get scheduled.
	BackupTemplate BackupSpec `json:"backupTemplate"`
}

// BackupSchedulePhase is a string representation of the lifecycle phase
// of a backup schedule.
type BackupSchedulePhase string

const (
	// BackupSchedulePhaseNew means the backup schedule has been created but not
	// yet processed by the backup schedule controller.
	BackupSchedulePhaseNew BackupSchedulePhase = "New"

	// BackupSchedulePhaseEnabled means the backup schedule has been validated and
	// will now be triggering backups according to the schedule spec.
	BackupSchedulePhaseEnabled BackupSchedulePhase = "Enabled"

	// BackupSchedulePhaseFailedValidation means the backup schedule has failed
	// the controller's validations and therefore will not trigger backups.
	BackupSchedulePhaseFailedValidation BackupSchedulePhase = "FailedValidation"
)

// ScheduleStatus captures the current state of a MySQL backup schedule.
type ScheduleStatus struct {
	// Phase is the current phase of the MySQL backup schedule.
	Phase BackupSchedulePhase `json:"phase"`

	// LastBackup is the last time a Backup was run for this
	// backup schedule.
	LastBackup metav1.Time `json:"lastBackup"`
}

// +genclient
// +genclient:noStatus

// BackupSchedule is a MySQL Operator resource that represents a backup
// schedule of a MySQL cluster.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BackupSchedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   BackupScheduleSpec `json:"spec"`
	Status ScheduleStatus     `json:"status,omitempty"`
}

// BackupScheduleList is a list of BackupSchedules.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BackupScheduleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []BackupSchedule `json:"items"`
}

// EnsureDefaults can be invoked to ensure the default values are present.
func (b BackupSchedule) EnsureDefaults() *BackupSchedule {
	buildVersion := version.GetBuildVersion()
	if buildVersion != "" {
		if b.Labels == nil {
			b.Labels = make(map[string]string)
		}
		_, hasKey := b.Labels[constants.MySQLOperatorVersionLabel]
		if !hasKey {
			SetOperatorVersionLabel(b.Labels, buildVersion)
		}
	}
	return &b
}

// Validate checks if the resource spec is valid.
func (b BackupSchedule) Validate() error {
	return validateBackupSchedule(&b).ToAggregate()
}

// RestoreSpec defines the specification for a restore of a MySQL backup.
type RestoreSpec struct {
	// ClusterRef is a refeference to the Cluster to which the Restore
	// belongs.
	ClusterRef *corev1.LocalObjectReference `json:"clusterRef"`

	// BackupRef is a reference to the Backup object to be restored.
	BackupRef *corev1.LocalObjectReference `json:"backupRef"`

	// AgentScheduled is the agent hostname to run the backup on
	AgentScheduled string `json:"agentscheduled"`
}

// RestorePhase represents the current life-cycle phase of a Restore.
type RestorePhase string

const (
	// RestorePhaseUnknown means that the restore hasn't yet been processed.
	RestorePhaseUnknown RestorePhase = ""

	// RestorePhaseNew means that the restore hasn't yet been processed.
	RestorePhaseNew RestorePhase = "New"

	// RestorePhaseScheduled means that the restore has been scheduled on an
	// appropriate replica.
	RestorePhaseScheduled RestorePhase = "Scheduled"

	// RestorePhaseStarted means the restore is in progress.
	RestorePhaseStarted RestorePhase = "Started"

	// RestorePhaseComplete means the restore has terminated successfully.
	RestorePhaseComplete RestorePhase = "Complete"

	// RestorePhaseFailed means the Restore has terminated with an error.
	RestorePhaseFailed RestorePhase = "Failed"
)

// RestoreStatus captures the current status of a MySQL restore.
type RestoreStatus struct {
	// Phase is the current life-cycle phase of the Restore.
	Phase RestorePhase `json:"phase"`

	// TimeStarted is the time at which the restore was started.
	TimeStarted metav1.Time `json:"timeStarted"`

	// TimeCompleted is the time at which the restore completed.
	TimeCompleted metav1.Time `json:"timeCompleted"`
}

// +genclient
// +genclient:noStatus

// Restore is a MySQL Operator resource that represents the restoration of
// backup of a MySQL cluster.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Restore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   RestoreSpec   `json:"spec"`
	Status RestoreStatus `json:"status"`
}

// RestoreList is a list of Restores.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Restore `json:"items"`
}

// EnsureDefaults can be invoked to ensure the default values are present.
func (r Restore) EnsureDefaults() *Restore {
	buildVersion := version.GetBuildVersion()
	if buildVersion != "" {
		if r.Labels == nil {
			r.Labels = make(map[string]string)
		}
		_, hasKey := r.Labels[constants.MySQLOperatorVersionLabel]
		if !hasKey {
			SetOperatorVersionLabel(r.Labels, buildVersion)
		}
	}
	return &r
}

// Validate checks if the resource spec is valid.
func (r Restore) Validate() error {
	return validateRestore(&r).ToAggregate()
}
