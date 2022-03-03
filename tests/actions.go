package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"

	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
)

type SendTokensAction struct {
	chain  uint
	from   uint
	to     uint
	amount uint
}

func (s System) sendTokens(action SendTokensAction) {
	bz, err := exec.Command("docker", "exec", s.containerConfig.instanceName, s.containerConfig.binaryName,

		"tx", "bank", "send",
		s.validatorConfigs[action.from].delAddress,
		s.validatorConfigs[action.to].delAddress,
		fmt.Sprint(action.amount)+`stake`,

		`--chain-id`, s.chainConfigs[action.chain].chainId,
		`--home`, s.getTxValidatorHome(action.chain, action.from),
		`--node`, "tcp://"+s.chainConfigs[action.chain].ipPrefix+".0:26658",
		`--keyring-backend`, `test`,
		`-b`, `block`,
		`-y`,
	).CombinedOutput()

	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}
}

type StartChainAction struct {
	chain          uint
	validators     []uint
	genesisChanges string
	skipGentx      bool
	copyConfigs    string
}

func (s System) startChain(
	action StartChainAction,
) {
	chainConfig := s.chainConfigs[action.chain]
	type jsonValAttrs struct {
		Mnemonic   string `json:"mnemonic"`
		Allocation string `json:"allocation"`
		Stake      string `json:"stake"`
		Number     string `json:"number"`
	}

	var validators []jsonValAttrs
	for _, val := range action.validators {
		validators = append(validators, jsonValAttrs{
			Mnemonic:   s.validatorConfigs[val].mnemonic,
			Allocation: chainConfig.initialAllocation,
			Stake:      chainConfig.stakeAmount,
			Number:     fmt.Sprint(val),
		})
	}

	vals, err := json.Marshal(validators)
	if err != nil {
		log.Fatal(err)
	}

	var genesisChanges string
	if action.genesisChanges != "" {
		genesisChanges = chainConfig.genesisChanges + " | " + action.genesisChanges
	} else {
		genesisChanges = chainConfig.genesisChanges
	}

	cmd := exec.Command("docker", "exec", s.containerConfig.instanceName, "/bin/bash",
		"/testnet-scripts/start-chain/start-chain.sh", s.containerConfig.binaryName, string(vals),
		chainConfig.chainId, chainConfig.ipPrefix, genesisChanges,
		fmt.Sprint(action.skipGentx), action.copyConfigs)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(cmdReader)

	for scanner.Scan() {
		out := scanner.Text()
		// fmt.Println("startChain: " + out)
		if out == "done!!!!!!!!" {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	s.addChainToRelayer(AddChainToRelayerAction{
		chain:     action.chain,
		validator: action.validators[0],
	})
}

type SubmitTextProposalAction struct {
	chain       uint
	from        uint
	deposit     uint
	propType    string
	title       string
	description string
}

func (s System) submitTextProposal(
	action SubmitTextProposalAction,
) {
	// docker exec interchain-security-instance interchain-securityd tx gov submit-proposal --title="Test Proposal" --description="My awesome proposal" --type Text --deposit 10000000stake --from validator1 --chain-id provider --home /provider/validator1 --keyring-backend test
	bz, err := exec.Command("docker", "exec", s.containerConfig.instanceName, s.containerConfig.binaryName,

		"tx", "gov", "submit-proposal",
		`--title`, action.title,
		`--description`, action.description,
		`--type`, action.propType,
		`--deposit`, fmt.Sprint(action.deposit)+`stake`,

		`--from`, `validator`+fmt.Sprint(action.from),
		`--chain-id`, s.chainConfigs[action.chain].chainId,
		`--home`, `/provider/validator`+fmt.Sprint(action.from),
		`--node`, "tcp://"+s.chainConfigs[action.chain].ipPrefix+".0:26658",
		`--keyring-backend`, `test`,
		`-b`, `block`,
		`-y`,
	).CombinedOutput()

	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}
}

type SubmitConsumerProposalAction struct {
	chain         uint
	from          uint
	deposit       uint
	consumerChain uint
	spawnTime     time.Time
	initialHeight clienttypes.Height
}

// TODO: import this directly from the module once it is merged
type CreateChildChainProposalJSON struct {
	Title         string             `json:"title"`
	Description   string             `json:"description"`
	ChainId       string             `json:"chain_id"`
	InitialHeight clienttypes.Height `json:"initial_height"`
	GenesisHash   []byte             `json:"genesis_hash"`
	BinaryHash    []byte             `json:"binary_hash"`
	SpawnTime     time.Time          `json:"spawn_time"`
	Deposit       string             `json:"deposit"`
}

func (s System) submitConsumerProposal(
	action SubmitConsumerProposalAction,
) {
	prop := CreateChildChainProposalJSON{
		Title:         "Create a chain",
		Description:   "Gonna be a great chain",
		ChainId:       s.chainConfigs[action.consumerChain].chainId,
		InitialHeight: action.initialHeight,
		GenesisHash:   []byte("gen_hash"),
		BinaryHash:    []byte("bin_hash"),
		SpawnTime:     action.spawnTime,
		Deposit:       fmt.Sprint(action.deposit) + `stake`,
	}

	bz, err := json.Marshal(prop)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: cleanup file
	bz, err = exec.Command("docker", "exec", s.containerConfig.instanceName, "/bin/bash", "-c", fmt.Sprintf(`echo '%s' > %s`, string(bz), "/temp-proposal.json")).CombinedOutput()

	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}

	bz, err = exec.Command("docker", "exec", s.containerConfig.instanceName, s.containerConfig.binaryName,

		"tx", "gov", "submit-proposal", "create-child-chain",
		"/temp-proposal.json",

		`--from`, `validator`+fmt.Sprint(action.from),
		`--chain-id`, s.chainConfigs[action.chain].chainId,
		`--home`, `/provider/validator`+fmt.Sprint(action.from),
		`--node`, "tcp://"+s.chainConfigs[action.chain].ipPrefix+".0:26658",
		`--keyring-backend`, `test`,
		`-b`, `block`,
		`-y`,
	).CombinedOutput()

	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}
}

type VoteGovProposalAction struct {
	chain      uint
	from       []uint
	vote       []string
	propNumber uint
}

func (s System) voteGovProposal(
	action VoteGovProposalAction,
) {
	var wg sync.WaitGroup
	for i, val := range action.from {
		wg.Add(1)
		vote := action.vote[i]
		go func(val uint, vote string) {
			defer wg.Done()
			bz, err := exec.Command("docker", "exec", s.containerConfig.instanceName, s.containerConfig.binaryName,

				"tx", "gov", "vote",
				fmt.Sprint(action.propNumber), vote,

				`--from`, `validator`+fmt.Sprint(val),
				`--chain-id`, s.chainConfigs[action.chain].chainId,
				`--home`, `/provider/validator`+fmt.Sprint(val),
				`--node`, "tcp://"+s.chainConfigs[action.chain].ipPrefix+".0:26658",
				`--keyring-backend`, `test`,
				`-b`, `block`,
				`-y`,
			).CombinedOutput()

			if err != nil {
				log.Fatal(err, "\n", string(bz))
			}
		}(val, vote)
	}

	wg.Wait()
	time.Sleep(time.Duration(s.chainConfigs[action.chain].votingWaitTime) * time.Second)
}

type StartConsumerChainAction struct {
	consumerChain uint
	providerChain uint
	validators    []uint
}

func (s System) startConsumerChain(action StartConsumerChainAction) {
	bz, err := exec.Command("docker", "exec", s.containerConfig.instanceName, s.containerConfig.binaryName,

		"query", "parent", "child-genesis",
		s.chainConfigs[action.consumerChain].chainId,

		`--node`, "tcp://"+s.chainConfigs[action.providerChain].ipPrefix+".0:26658",
		`-o`, `json`,
	).CombinedOutput()

	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}

	s.startChain(StartChainAction{
		chain:          1,
		validators:     action.validators,
		genesisChanges: ".app_state.ccvchild = " + string(bz),
		skipGentx:      true,
		copyConfigs:    s.chainConfigs[action.providerChain].chainId,
	})
}

type AddChainToRelayerAction struct {
	chain     uint
	validator uint
}

const hermesChainConfigTemplate = `

[[chains]]
account_prefix = "cosmos"
clock_drift = "5s"
gas_adjustment = 0.1
grpc_addr = "%s"
id = "%s"
key_name = "%s"
max_gas = 2000000
rpc_addr = "%s"
rpc_timeout = "10s"
store_prefix = "ibc"
trusting_period = "14days"
websocket_addr = "%s"

[chains.gas_price]
	denom = "stake"
	price = 0.001

[chains.trust_threshold]
	denominator = "3"
	numerator = "1"
`

func (s System) addChainToRelayer(action AddChainToRelayerAction) {
	valIp := s.chainConfigs[action.chain].ipPrefix + `.` + fmt.Sprint(action.validator)
	chainId := s.chainConfigs[action.chain].chainId
	keyName := "validator" + fmt.Sprint(action.validator)
	rpcAddr := "http://" + valIp + ":26658"
	grpcAddr := "tcp://" + valIp + ":9091"
	wsAddr := "ws://" + valIp + ":26657/websocket"

	chainConfig := fmt.Sprintf(hermesChainConfigTemplate,
		grpcAddr,
		chainId,
		keyName,
		rpcAddr,
		wsAddr,
	)

	bashCommand := fmt.Sprintf(`echo '%s' >> %s`, chainConfig, "/root/.hermes/config.toml")

	bz, err := exec.Command("docker", "exec", s.containerConfig.instanceName, "bash", "-c",
		bashCommand,
	).CombinedOutput()
	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}

	bz, err = exec.Command("docker", "exec", s.containerConfig.instanceName, "/root/.cargo/bin/hermes",
		"keys", "restore",
		"--mnemonic", s.validatorConfigs[action.validator].mnemonic,
		s.chainConfigs[action.chain].chainId,
	).CombinedOutput()

	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}
}

type AddIbcConnectionAction struct {
	chainA  uint
	chainB  uint
	clientA uint
	clientB uint
	order   string
}

func (s System) addIbcConnection(action AddIbcConnectionAction) {
	cmd := exec.Command("docker", "exec", s.containerConfig.instanceName, "/root/.cargo/bin/hermes",
		"create", "connection",
		s.chainConfigs[action.chainA].chainId,
		"--client-a", "07-tendermint-"+fmt.Sprint(action.clientA),
		"--client-b", "07-tendermint-"+fmt.Sprint(action.clientB),
	)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(cmdReader)

	for scanner.Scan() {
		out := scanner.Text()
		// fmt.Println("addIbcConnection: " + out)
		if out == "done!!!!!!!!" {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

type AddIbcChannelAction struct {
	chainA      uint
	chainB      uint
	connectionA uint
	portA       string
	portB       string
	order       string
}

func (s System) addIbcChannel(action AddIbcChannelAction) {
	// // hermes create channel ibc-1 ibc-2 --port-a transfer --port-b transfer -o unordered
	cmd := exec.Command("docker", "exec", s.containerConfig.instanceName, "/root/.cargo/bin/hermes",
		"create", "channel",
		s.chainConfigs[action.chainA].chainId,
		"--port-a", action.portA,
		"--port-b", action.portB,
		"-o", action.order,
		"--channel-version", s.containerConfig.ccvVersion,
		"--connection-a", "connection-"+fmt.Sprint(action.connectionA),
	)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(cmdReader)

	for scanner.Scan() {
		out := scanner.Text()
		// fmt.Println("addIBCChannel: " + out)
		if out == "done!!!!!!!!" {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

type RelayPacketsAction struct {
	chain     uint
	port      string
	channelId string
}

func (s System) relayPackets(action RelayPacketsAction) {
	// hermes clear packets ibc0 transfer channel-13
	bz, err := exec.Command("docker", "exec", s.containerConfig.instanceName, "$HOME/.cargo/bin/hermes", "clear", "packets",
		s.chainConfigs[action.chain].chainId, action.port, action.channelId,
	).CombinedOutput()

	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}
}

func (s System) getQueryValidatorHome(chain uint) string {
	// Get first subdirectory of the directory of this chain, which will be the home directory of one of the validators
	bz, err := exec.Command("docker", "exec", s.containerConfig.instanceName, "bash", "-c", `cd /`+s.chainConfigs[chain].chainId+`; ls -d */ | awk '{print $1}' | head -n 1`).CombinedOutput()

	if err != nil {
		log.Fatal(err, "\n", string(bz))
	}

	return `/` + s.chainConfigs[chain].chainId + `/` + string(bz)
}

func (s System) getTxValidatorHome(chain uint, validator uint) string {
	return `/` + s.chainConfigs[chain].chainId + `/validator` + fmt.Sprint(validator)
}
