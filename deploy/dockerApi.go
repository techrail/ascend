package deploy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/techrail/ascend/constants"
	"github.com/techrail/ascend/models"
)

func GetContext(filePath string) io.Reader {
	ctx, _ := archive.TarWithOptions(filePath, &archive.TarOptions{})
	return ctx
}

func DockerAPI(deployRequest models.DeployRequest, response chan models.DockerResponse) {
	err := validateRequest(deployRequest)

	if err != nil {
		response <- models.DockerResponse{Error: stringPtr(fmt.Sprintf("%v", err))}
		return
	}

	ctx := context.Background()
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
		return
	}
	defer dockerClient.Close()

	imageName := namesgenerator.GetRandomName(0)

	buildImage(dockerClient, deployRequest, imageName)

	dockerImageInfo := startContainer(dockerClient, ctx, deployRequest, imageName)
	response <- models.DockerResponse{Info: &dockerImageInfo}
}

func buildImage(dockerClient *client.Client, deployRequest models.DeployRequest, imageName string) {
	dockerBuildContext := GetContext("./Dockerfile")
	buildArgs := make(map[string]*string)

	executableName := extractExecutableNameFromBuildCommand(deployRequest.BuildCommand)

	buildArgs["GIT_URL"] = deployRequest.RepositoryUrl
	buildArgs["PORT"] = deployRequest.Port
	buildArgs["BUILD_CMD"] = deployRequest.BuildCommand
	buildArgs["START_CMD"] = deployRequest.StartCommand
	buildArgs["EXEC_NAME"] = &executableName
	buildOptions := types.ImageBuildOptions{
		Tags:      []string{imageName},
		Remove:    true,
		BuildArgs: buildArgs,
		NoCache:   true,
	}
	log.Printf("I#23MJZ0 - Docker build started for image: %v\n", imageName)
	buildResponse, err := dockerClient.ImageBuild(context.Background(), dockerBuildContext, buildOptions)
	if err != nil {
		log.Fatal(err)
	}

	createLogsDirectory()

	file, err := os.Create(fmt.Sprintf("%v/%v_ImageBuild_%v.log", constants.ContainerLogsDirectory, imageName, time.Now().Format("20060102150405")))

	if err != nil {
		log.Panicln("P#23M32K - ", err)
	}

	io.Copy(file, buildResponse.Body)

	defer buildResponse.Body.Close()
}

func startContainer(dockerClient *client.Client, ctx context.Context, deployRequest models.DeployRequest, imageName string) models.DeploymentInfo {
	portBindings := make(nat.PortMap)
	var sb strings.Builder
	if deployRequest.Port != nil {
		sb.WriteString(*deployRequest.Port)
		sb.WriteString("/tcp")
	} else {
		sb.WriteString(constants.DockerDefaultProtocolAndPort)
	}

	containerPort := sb.String()

	hostPort, err := GetFreePort()

	if err != nil {
		log.Fatal(err)
	}

	bindings := []nat.PortBinding{
		{HostIP: "", HostPort: hostPort},
	}
	portBindings[nat.Port(containerPort)] = bindings

	resources := setContainerResources(deployRequest)

	mounts := setContainerMounts(deployRequest)

	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image: imageName,
	}, &container.HostConfig{
		PortBindings: portBindings,
		NetworkMode:  constants.DockerDefaultNetworkMode,
		Resources:    *resources,
		Mounts:       mounts,
	}, nil, nil, "")

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("I#23MK1B - Container started for image : %v\n", imageName)
	if err := dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		log.Fatal(err)
	}

	statusCh, errCh := dockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNextExit)
	out, err1 := dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, Follow: true})
	if err1 != nil {
		log.Panic(err1)
	}

	file, err := os.Create(fmt.Sprintf("%v/%v_%v.log", constants.ContainerLogsDirectory, imageName, time.Now().Format("20060102150405")))
	if err != nil {
		log.Panicln("P#23M32K - ", err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			log.Panic(err)
		}
	case status := <-statusCh:
		log.Printf("I#23M3DC - %v", status.StatusCode)
	default:
		go stdcopy.StdCopy(file, os.Stderr, out)
	}

	return models.DeploymentInfo{ImageName: imageName, Port: hostPort}
}

// GetFreePort asks the kernel for a free open port that is ready to use.
func GetFreePort() (port string, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return strconv.Itoa(l.Addr().(*net.TCPAddr).Port), nil
		}
	}
	return
}

func extractExecutableNameFromBuildCommand(buildCommand *string) string {
	commands := strings.Fields(*buildCommand)

	for i, part := range commands {
		if part == "-o" && i+1 < len(commands) {
			return commands[i+1]
		}
	}
	return constants.GoDefaultExecutableName
}

func setContainerResources(deployRequest models.DeployRequest) *container.Resources {
	resources := container.Resources{}

	if deployRequest.MemoryLimit != nil {
		resources.Memory = *deployRequest.MemoryLimit
	} else {
		resources.Memory = constants.DockerContainerMemoryLimit
	}

	if deployRequest.CPUs != nil {
		resources.NanoCPUs = int64(*deployRequest.CPUs * float64(1e+9))
	}

	return &resources
}

func setContainerMounts(deployRequest models.DeployRequest) []mount.Mount {
	if deployRequest.Mounts == nil {
		return nil
	}

	var mounts []mount.Mount

	for _, v := range *deployRequest.Mounts {
		var m mount.Mount
		m.Source = v.Source
		m.Target = v.Target
		m.Type = getMountType(v.Type)
		mounts = append(mounts, m)
	}

	return mounts
}

func getMountType(mt string) mount.Type {
	switch strings.ToLower(mt) {
	case "bind":
		return mount.TypeBind
	case "volume":
		return mount.TypeVolume
	case "cluster":
		return mount.TypeCluster
	case "namedpipe":
		return mount.TypeNamedPipe
	case "tmpfs":
		return mount.TypeTmpfs
	default:
		return mount.TypeBind
	}
}

func createLogsDirectory() {
	if _, err := os.Stat(constants.ContainerLogsDirectory); os.IsNotExist(err) {
		err := os.Mkdir(constants.ContainerLogsDirectory, 0777)
		if err != nil {
			log.Panicln("P#23M30E - ", err)
		}
	}
}

func validateRequest(deployRequest models.DeployRequest) error {
	if deployRequest.RepositoryUrl == nil {
		return fmt.Errorf("repository url not provided")
	}

	if deployRequest.MemoryLimit != nil && *deployRequest.MemoryLimit < int64(constants.DockerContainerMinMemoryLimit) {
		return fmt.Errorf("minimum memory limit allowed is 6mb")
	}

	if deployRequest.CPUs != nil && *deployRequest.CPUs < constants.DockerContainerMinCPULimit {
		return fmt.Errorf("minimum cpu limit allowed is 0.01")
	}

	return nil
}

func stringPtr(s string) *string {
	return &s
}
