package core

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/algorand/go-algorand/config"
	generatedV2 "github.com/algorand/go-algorand/daemon/algod/api/server/v2/generated"
	"github.com/algorand/go-algorand/netdeploy"
	"github.com/algorand/go-algorand/nodecontrol"
)

//go:embed OneNode.json
var networkTemplate string

// NetworkData is stored for each active network.
type NetworkData struct {
	bindir     string
	privatenet netdeploy.Network
}

// Network maintains state for the networks being controlled by goran.
type Network struct {
	networks map[string]*NetworkData

	defaultBinDir   string
	defaultTmpDir   string
	defaultTemplate string
}

// MakeNetwork initializes the goran Network object.
func MakeNetwork(binDir, tmpDir string) (*Network, error) {
	// Initialize default network template
	templatePath := path.Join(tmpDir, "OneNode.json")
	f, err := os.Create(templatePath)
	if os.IsExist(err) {
		// done, it's there already
	} else if err != nil {
		return nil, fmt.Errorf("MakeNetwork() failed to write template: %w", err)
	} else {
		f.WriteString(networkTemplate)
	}

	return &Network{
		networks:        make(map[string]*NetworkData),
		defaultBinDir:   binDir,
		defaultTmpDir:   tmpDir,
		defaultTemplate: templatePath,
	}, nil
}

// NetworkStartArgs args for the Network.Start function.
type NetworkStartArgs struct {
	BinDir       string // optional
	TmpDir       string // optional
	TemplateFile string // optional
	DevMode      bool   // optional
	Name         string
}

// NetworkStartReply reply from the Network.Start function.
type NetworkStartReply = NetworkStatusReply

// Start a private network.
func (s *Network) Start(args *NetworkStartArgs, reply *NetworkStartReply) error {
	if args.Name != "" {
		if _, ok := s.networks[args.Name]; ok {
			return fmt.Errorf("Start() duplicate name")
		}
	}

	// If no name is provided, provide a default.
	if args.Name == "" {
		args.Name = fmt.Sprintf("%d", len(s.networks))
	}

	// If no binary directory override is provided, use the default.
	if args.BinDir == "" {
		args.BinDir = s.defaultBinDir
	}

	// If no temp directory override is provided, use the default.
	if args.TmpDir == "" {
		args.TmpDir = s.defaultTmpDir
	}

	// If no temp directory override is provided, use the default.
	if args.TemplateFile == "" {
		args.TemplateFile = s.defaultTemplate
	}

	// Initialize the new network directory.
	dataDir := path.Join(args.TmpDir, args.Name)
	if err := os.Mkdir(dataDir, 0755); err != nil {
		return fmt.Errorf("Start() err: %w", err)
	}

	var callback nodecontrol.AlgodExitErrorCallback = nil
	var consensus config.ConsensusProtocols
	importKeys := true
	network, err := netdeploy.CreateNetworkFromTemplate(args.Name, dataDir, args.TemplateFile, args.BinDir, importKeys, callback, consensus, args.DevMode)
	if err != nil {
		return fmt.Errorf("Start() failed to create from template: %w", err)
	}

	// Start the private network.
	err = network.Start(args.BinDir, true)
	if err != nil {
		return fmt.Errorf("Start() failed to start: %w", err)
	}

	state := &NetworkData{
		privatenet: network,
		bindir:     args.BinDir,
	}

	s.networks[args.Name] = state

	// Return status for the new network.
	return s.Status(&NetworkStatusArgs{Name: args.Name}, reply)
}

// NetworkStopArgs args for the Network.Stop function.
type NetworkStopArgs struct {
	Name string
}

// NetworkStopReply reply from the Network.Stop function.
type NetworkStopReply struct {
	Name string
}

// Stop a running private network known by the system.
func (s *Network) Stop(args *NetworkStopArgs, reply *NetworkStopReply) error {
	network, ok := s.networks[args.Name]
	if !ok {
		return fmt.Errorf("Stop(): unknown network '%s'", args.Name)
	}

	network.privatenet.Stop(network.bindir)

	reply.Name = args.Name
	return nil
}

// NetworkStatusArgs args for the Network.Status function.
type NetworkStatusArgs struct {
	Name string
}

// NetworkStatusReply reply from the Network.Status function.
type NetworkStatusReply generatedV2.NodeStatusResponse

// Status of a private network.
func (s *Network) Status(args *NetworkStatusArgs, reply *NetworkStatusReply) error {
	network, ok := s.networks[args.Name]
	if !ok {
		return fmt.Errorf("Status(): unknown network '%s'", args.Name)
	}

	replies := network.privatenet.NodesStatus(network.bindir)
	if len(replies) != 1 {
		var buf strings.Builder
		sep := ""
		for k, _ := range replies {
			buf.WriteString(sep)
			buf.WriteString(k)
			sep = ", "
		}
		return fmt.Errorf("Status(): unexpected number of replies: %s", buf.String())
	}

	for _, v := range replies {
		if v.Error != nil {
			return fmt.Errorf("Status(): network returned an error: %s", v.Error)
		}
		*reply = NetworkStatusReply(v.Status)
	}
	return nil
}
