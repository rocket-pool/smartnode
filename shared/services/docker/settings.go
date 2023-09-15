package docker

// Docker container names belonging to Rocket Pool
type ContainerName string

// Container name items
const (
	ContainerName_ExecutionClient ContainerName = "eth1"
	ContainerName_BeaconNode      ContainerName = "eth2"
	ContainerName_ValidatorClient ContainerName = "validator"
	ContainerName_MevBoost        ContainerName = "mev-boost"
	ContainerName_Node            ContainerName = "node"
)

// Settings
const (
	DockerApiVersion string = "1.40"
)
