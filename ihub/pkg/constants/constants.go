package constants

// Context Parameters
const Destination = "Destination"
const ClusterName = "ClusterName"
const ClusterDomain = "ClusterDomain"
const NeedApprove = "NeedApprove"

// const Role = "Role"

// Destination
const DestinationIn = "in"
const DestinationOut = "out"

// Role
const RoleClusterAdmin = 0
const RoleGroupAdmin = 1
const RoleUser = 2

// Runmode
const RunmodeOut = "out"
const RunmodeIn = "in"

// ClusterStatus
const ClusterStatusReseting = 2
const ClusterStatusResetSucceed = 1
const ClusterStatusResetFailed = 0

// HTTP Header variables
const (
	HTTPHeaderClusterName = "X-Cluster-Name"
	HTTPHeaderTraceID     = "X-Trace-ID"
)

// Default value for rgm
const (
	DefaultLogName = "ihub.log"
)
