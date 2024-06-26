package docker

import (
	"errors"
	"time"

	myTypes "github.com/kekaadrenalin/dockhook/pkg/types"
	log "github.com/sirupsen/logrus"
)

func CreateClients(args myTypes.Args) map[string]myTypes.Client {
	clients := createClients(args, NewClientWithFilters, NewClientWithTLSAndFilter, args.Hostname)

	if len(clients) == 0 {
		log.Fatal("Could not connect to any Docker Engines")
	}

	log.Infof("Connected to %d Docker Engine(s)", len(clients))

	return clients
}

func createClients(
	args myTypes.Args,
	localClientFactory func(map[string][]string) (myTypes.Client, error),
	remoteClientFactory func(map[string][]string, myTypes.Host) (myTypes.Client, error),
	hostname string,
) map[string]myTypes.Client {
	clients := make(map[string]myTypes.Client)

	if localClient, err := createLocalClient(args, localClientFactory); err == nil {
		if hostname != "" {
			localClient.Host().Name = hostname
		}

		clients[localClient.Host().ID] = localClient
	}

	for _, remoteHost := range args.RemoteHost {
		host, err := ParseConnection(remoteHost)
		if err != nil {
			log.Fatalf("Could not parse remote host %s: %s", remoteHost, err)
		}

		log.Debugf("Creating remote client for %s with %+v", host.Name, host)
		log.Infof("Creating client for %s with %s", host.Name, host.URL.String())

		if client, err := remoteClientFactory(args.Filter, host); err == nil {
			if _, err := client.ListContainers(); err == nil {
				log.Debugf("Connected to local Docker Engine")
				clients[client.Host().ID] = client
			} else {
				log.Warningf("Could not connect to remote host %s: %s", host.ID, err)
			}
		} else {
			log.Warningf("Could not create client for %s: %s", host.ID, err)
		}
	}

	return clients
}

func createLocalClient(
	args myTypes.Args,
	localClientFactory func(map[string][]string) (myTypes.Client, error),
) (myTypes.Client, error) {
	for i := 1; ; i++ {
		dockerClient, err := localClientFactory(args.Filter)
		if err == nil {
			_, err := dockerClient.ListContainers()
			if err != nil {
				log.Debugf("Could not connect to local Docker Engine: %s", err)
			} else {
				log.Debugf("Connected to local Docker Engine")
				return dockerClient, nil
			}
		}

		if args.WaitForDockerSeconds > 0 {
			log.Infof("Waiting for Docker Engine (attempt %d): %s", i, err)
			time.Sleep(5 * time.Second)
			args.WaitForDockerSeconds -= 5
		} else {
			log.Debugf("Local Docker Engine not found")
			break
		}
	}

	return nil, errors.New("could not connect to local Docker Engine")
}
