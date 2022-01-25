package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
)

// property erupt day common remind oblige chunk thumb jazz camera erupt reward divorce fit toy cargo traffic scrub begin gown recall video friend prosper
// decide praise business actor peasant farm drastic weather extend front hurt later song give verb rhythm worry fun pond reform school tumble august one
// brown include source lesson joy fringe great hazard breeze essay hurdle gadget make prepare unfair sense divorce emotion double elite more subway hat worth
// sight similar better jar bitter laptop solve fashion father jelly scissors chest uniform play unhappy convince silly clump another conduct behave reunion marble animal
// glass trip produce surprise diamond spin excess gaze wash drum human solve dress minor artefact canoe hard ivory orange dinner hybrid moral potato jewel
// pave immune ethics wrap gain ceiling always holiday employ earth tumble real ice engage false unable carbon equal fresh sick tattoo nature pupil nuclear

// Attributes that are unique to a validator. Allows us to map (part of) the set of uints to
// a set of viable validators
type ValidatorKeys struct {
	mnemonic   string
	delAddress string
	valAddress string
}

// Attributes that are unique to a chain. Allows us to map (part of) the set of uints to
// a set of viable chains
type ChainAttrs struct {
	chainId        string
	ipPrefix       string
	genesisChanges string
	rpcPort        uint
	grpcPort       uint
}

// These values will not be altered during a typical test run
// They are probably not part of the model
type Config struct {
	containerName     string
	instanceName      string
	binaryName        string
	exposePorts       []uint
	startChainScript  string
	initialAllocation string
	stakeAmount       string
	validatorsKeys    []ValidatorKeys
	chainAttrs        []ChainAttrs
}

func DefautlSystemConfig() Config {
	return Config{
		containerName:     "interchain-security-container",
		instanceName:      "interchain-security-instance",
		binaryName:        "interchain-securityd",
		exposePorts:       []uint{9090, 26657, 9089, 26656},
		startChainScript:  "/testnet-scripts/start-chain/start-chain.sh",
		initialAllocation: "10000000000stake,10000000000footoken",
		stakeAmount:       "500000000stake",
		validatorsKeys: []ValidatorKeys{
			{
				mnemonic: "pave immune ethics wrap gain ceiling always holiday employ earth tumble real ice engage false unable carbon equal fresh sick tattoo nature pupil nuclear",
			},
			{
				mnemonic: "glass trip produce surprise diamond spin excess gaze wash drum human solve dress minor artefact canoe hard ivory orange dinner hybrid moral potato jewel",
			},
			{
				mnemonic: "sight similar better jar bitter laptop solve fashion father jelly scissors chest uniform play unhappy convince silly clump another conduct behave reunion marble animal",
			},
		},
		chainAttrs: []ChainAttrs{
			{
				chainId:        "provider",
				ipPrefix:       "7.7.7",
				genesisChanges: ".app_state.gov.voting_params.voting_period = \"60s\"",
				rpcPort:        26657,
				grpcPort:       9090,
			},
		},
	}
}

type State struct {
	transferSent bool
}

type System struct {
	config Config
	state  State
}

func main() {
	s := System{
		config: DefautlSystemConfig(),
	}

	s.startDocker()
	println("docker started?")

	s.startChain(0, []uint{0, 1, 2})
	println("chain started?")

	bz, _ := exec.Command("docker", "exec", "interchain-security-instance", "interchain-securityd", "query", "bank", "balances",
		`cosmos19pe9pg5dv9k5fzgzmsrgnw9rl9asf7ddwhu7lm`,
		`--chain-id`, `provider`,
		`--home`, `/provider/validator1`,
	).CombinedOutput()
	fmt.Println(string(bz))

	// docker exec interchain-security-instance interchain-securityd tx bank send cosmos19pe9pg5dv9k5fzgzmsrgnw9rl9asf7ddwhu7lm cosmos1dkas8mu4kyhl5jrh4nzvm65qz588hy9qcz08la 1stake --home /provider/validator1 --keyring-backend test --chain-id provider -y
	bz, _ = exec.Command("docker", "exec", "interchain-security-instance", "interchain-securityd", "tx", "bank", "send",
		`cosmos19pe9pg5dv9k5fzgzmsrgnw9rl9asf7ddwhu7lm`,
		`cosmos1dkas8mu4kyhl5jrh4nzvm65qz588hy9qcz08la`,
		`1stake`,
		`--chain-id`, `provider`,
		`--home`, `/provider/validator1`,
		`--keyring-backend`, `test`,
		`-b`, `block`,
		`-y`,
	).CombinedOutput()

	fmt.Println(string(bz))

	// docker exec interchain-security-instance interchain-securityd query bank balances cosmos19pe9pg5dv9k5fzgzmsrgnw9rl9asf7ddwhu7lm --home /provider/validator1 --chain-id provider
	bz, _ = exec.Command("docker", "exec", "interchain-security-instance", "interchain-securityd", "query", "bank", "balances",
		`cosmos19pe9pg5dv9k5fzgzmsrgnw9rl9asf7ddwhu7lm`,
		`--chain-id`, `provider`,
		`--home`, `/provider/validator1`,
	).CombinedOutput()

	fmt.Println(string(bz))
}

func (s System) startDocker() {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	ports_string := ""

	for _, port := range s.config.exposePorts {
		ports_string = ports_string + " -p " + fmt.Sprint(port) + ":" + fmt.Sprint(port)
	}

	cmd := exec.Command("/bin/bash", path+"/start-docker.sh", s.config.containerName, s.config.instanceName, ports_string)

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
		// fmt.Println("startDocker: " + out)
		if out == "beacon!!!!!!!!!!" {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (s System) startChain(
	chain uint,
	validators []uint,
) {
	c := s.config.chainAttrs[chain]
	var mnemonics []string

	for _, val := range validators {
		mnemonics = append(mnemonics, s.config.validatorsKeys[val].mnemonic)
	}

	mnz, err := json.Marshal(mnemonics)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("docker", "exec", s.config.instanceName, "/bin/bash",
		s.config.startChainScript, s.config.binaryName, string(mnz), c.chainId, c.ipPrefix,
		fmt.Sprint(c.rpcPort), fmt.Sprint(c.grpcPort), c.genesisChanges, s.config.initialAllocation, s.config.stakeAmount)

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
			return
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
