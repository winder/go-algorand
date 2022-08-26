// Copyright (C) 2019-2022 Algorand, Inc.
// This file is part of go-algorand
//
// go-algorand is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-algorand is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with go-algorand.  If not, see <https://www.gnu.org/licenses/>.

package core

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/algorand/go-algorand/config"
	generatedV2 "github.com/algorand/go-algorand/daemon/algod/api/server/v2/generated"
	"github.com/algorand/go-algorand/netdeploy"
	"github.com/algorand/go-algorand/nodecontrol"
	"github.com/algorand/go-algorand/util"
	"github.com/algorand/go-algorand/util/tokens"
)

//go:embed OneNode.json
var networkTemplate string

// NetworkData is stored for each active network.
type NetworkData struct {
	rootDir    string
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

const metadataFile = ".goran"

type goranMetadata struct {
	BinDir string
}

func createNetwork(args *NetworkStartArgs, rootDir string) (*NetworkData, error) {
	var callback nodecontrol.AlgodExitErrorCallback = nil
	var consensus config.ConsensusProtocols
	importKeys := true
	network, err := netdeploy.CreateNetworkFromTemplate(args.Name, rootDir, args.TemplateFile, args.BinDir, importKeys, callback, consensus, args.DevMode)
	if err != nil {
		return nil, fmt.Errorf("createNetwork() err: %w", err)
	}

	metadata := goranMetadata{
		BinDir: args.BinDir,
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("initializeNetworks() failed to marshal metadata file: %w", err)
	}
	os.WriteFile(path.Join(rootDir, metadataFile), data, 0755)

	return &NetworkData{
		privatenet: network,
		bindir:     args.BinDir,
		rootDir:    rootDir,
	}, nil
}

func loadNetwork(rootDir string) (*NetworkData, error) {
	result := &NetworkData{}
	var err error
	// TODO: LoadNetwork does not work as expected.
	result.privatenet, err = netdeploy.LoadNetwork(rootDir)
	if err != nil {
		return nil, fmt.Errorf("loadNetwork() failed to load network: %w", err)
	}

	// Load properties from '.goran' metadata file.
	data, err := os.ReadFile(path.Join(rootDir, ".goran"))
	if err != nil {
		return nil, fmt.Errorf("loadNetwork() failed to read goran metadata file: %w", err)
	}

	var metadata goranMetadata
	if err != nil {
		return nil, fmt.Errorf("loadNetwork() failed to parse goran metadata file: %w", err)
	}
	err = json.Unmarshal(data, &metadata)

	return result, nil
}

func initializeNetworks(tmpDir string) (map[string]*NetworkData, error) {
	networks := make(map[string]*NetworkData)

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("initializeNetworks() failed to list temp dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			network, err := loadNetwork(entry.Name())
			if err != nil {
				return nil, fmt.Errorf("initializeNetworks() failed to load dir '%s': %w", entry.Name(), err)
			}
			networks[entry.Name()] = network
		}
	}

	return networks, nil
}

// MakeNetwork initializes the goran Network object.
// TODO: Should the tmpDir be scanned for existing networks?
// TODO: This could be the basis of a "snapshot" and "restore" feature.
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

	networks, err := initializeNetworks(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("MakeNetwork err: %w", err)
	}

	return &Network{
		networks:        networks,
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
type NetworkStartReply struct {
	Token      string
	AdminToken string
	URL        string
}

// Start a private network.
func (s *Network) Start(args NetworkStartArgs, reply *NetworkStartReply) error {
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

	var network *NetworkData

	// Initialize the new network directory.
	rootDir := path.Join(args.TmpDir, args.Name)
	err := os.Mkdir(rootDir, 0755)
	if os.IsExist(err) {
		network, err = loadNetwork(rootDir)
		// open existing network.
		if err != nil {
			return fmt.Errorf("Start() failed to open existing network '%s': %w", rootDir, err)
		}
	} else if err != nil {
		// error, return.
		return fmt.Errorf("Start() err: %w", err)
	} else {
		// no error, create a new network
		network, err = createNetwork(&args, rootDir)
		if err != nil {
			return fmt.Errorf("Start() failed to create network: %w", err)
		}
	}

	// Start the private network.
	err = network.privatenet.Start(args.BinDir, true)
	if err != nil {
		return fmt.Errorf("Start() failed to start: %w", err)
	}

	dataDir := network.privatenet.PrimaryDataDir()

	adminToken, err := tokens.GetAndValidateAPIToken(dataDir, tokens.AlgodAdminTokenFilename)
	if err != nil {
		network.privatenet.Stop(args.Name)
		return fmt.Errorf("Start() failed to lookup admin token: ", err)
	}

	token, err := tokens.GetAndValidateAPIToken(dataDir, tokens.AlgodTokenFilename)
	if err != nil {
		network.privatenet.Stop(args.Name)
		return fmt.Errorf("Start() failed to lookup token: ", err)
	}

	address, err := util.GetFirstLineFromFile(filepath.Join(dataDir, "algod.net"))
	if err != nil {
		network.privatenet.Stop(args.Name)
		return fmt.Errorf("Start() failed to lookup address: ", err)
	}

	s.networks[args.Name] = network

	*reply = NetworkStartReply{
		Token:      token,
		AdminToken: adminToken,
		URL:        address,
	}
	return nil
}

// NetworkDestroyArgs args for the Network.Destroy function.
type NetworkDestroyArgs struct {
	All  bool
	Name string
}

// NetworkDestroyReply reply from the Network.Destroy function.
type NetworkDestroyReply struct {
	Name string
}

// Destroy a running private network known by the system.
func (s *Network) Destroy(args NetworkDestroyArgs, reply *NetworkDestroyReply) error {
	if args.All {
		// TODO: destroy all networks.
	}

	network, ok := s.networks[args.Name]
	if !ok {
		return fmt.Errorf("Destroy(): unknown network '%s'", args.Name)
	}

	network.privatenet.Stop(network.bindir)
	err := os.RemoveAll(network.rootDir)
	if err != nil {
		fmt.Errorf("Destroy(): failed to remove dir '%s': %w", network.rootDir, err)
	}

	delete(s.networks, args.Name)

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
func (s *Network) Status(args NetworkStatusArgs, reply *NetworkStatusReply) error {
	network, ok := s.networks[args.Name]
	if !ok {
		return fmt.Errorf("Status(): unknown network '%s'", args.Name)
	}

	replies := network.privatenet.NodesStatus(network.bindir)
	if len(replies) != 1 {
		var buf strings.Builder
		sep := ""
		for k := range replies {
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

type NetworkListArgs = NetworkStatusArgs

type NetworkListReply struct {
	Networks []string
}

func (s *Network) List(_ NetworkListArgs, reply *NetworkListReply) error {
	networks := make([]string, 0, len(s.networks))
	for k := range s.networks {
		networks = append(networks, k)
	}
	reply.Networks = networks
	return nil
}
