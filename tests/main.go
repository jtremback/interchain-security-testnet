package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	start_docker("interchain-security-container", "interchain-security-instance", 9090, 26657, 1317, 8545)
	println("docker started?")

	// cmd := exec.Command("docker", "ps")
	// stdoutStderr, err := cmd.CombinedOutput()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("%s\n", stdoutStderr)

	start_chain("interchain-security-instance", "/testnet-scripts/start-chain/start-chain.sh", "interchain-securityd", 3, "provider", "7.7.7", 26657, 9090, ".app_state.gov.voting_params.voting_period = \"60s\"")

	mnemonic1 := strings.Split(catFileInDocker("interchain-security-instance", "/provider/validator1/mnemonic"), "\n")[5]

	println(mnemonic1)

	cmd := exec.Command("interchain-securityd", "keys", "add", "validator1", "--recover")
	cmd.Stdin = strings.NewReader(mnemonic1 + "\npassword\npassword")
	bz, err := cmd.CombinedOutput()

	fmt.Println(string(bz))

	if err != nil {
		log.Fatal(err)
	}

	// exec.Command("interchain-securityd", "tx", "gov", "submit-proposal", `--title="Test Proposal"`, `--description="My awesome proposal"`, `--type="Text"`, `--deposit="10test"`, "--from")

	// Pass gov prop
	// Query IBC module for client state and consensus state????? <- NO but should be possible
	// Instead, create client state:
	// clientState = ibctmtypes.NewClientState(<provider chain id>, ibctmtypes.DefaultTrustLevel, <unbonding period- query provider staking params>, <half of the previous argument>,
	// 		time.Second*10, <height at which the proposal passed on provider +-1>, commitmenttypes.GetSDKSpecs(), []string{"upgrade", "upgradedIBCState"}, true, true),
	// Then, create consensus state:
	// ConsensusState = ibctmtypes.NewConsensusState(<time at which the proposal passed on provider +-1>, commitmenttypes.NewMerkleRoot(<AppHash from block header from when the proposal passed on the provider +-1>), <NextValidatorsHash from block header from when the proposal passed on the provider +-1>)
	// Populate consumer genesis - set enabled to true
	// start consumer chain

	// # /bin/bash "$DIR/start-chain/start-chain.sh" interchain-securityd 3 consumer 7.7.8 26757 9190 '.app_state.gov.voting_params.voting_period = "60s"'

	// Set up relayer- Hermes may be best
	// https://hermes.informal.systems/config.html
	// - rpc endpoint
	// - private keys of an address with money
	// - probably mostly change id-key_name in config
	// - Massive headache: off by one errors

	println("chain started?")

	outChannel := make(chan string)
	_ = <-outChannel
}

func start_docker(container_name string, instance_name string, expose_ports ...uint) {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	ports_string := ""

	for _, port := range expose_ports {
		ports_string = ports_string + " -p " + fmt.Sprint(port) + ":" + fmt.Sprint(port)
	}

	cmd := exec.Command("/bin/bash", path+"/start-docker.sh", container_name, instance_name, ports_string)

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
		// fmt.Println("out: " + out)
		if out == "beacon!!!!!!!!!!" {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

// "/testnet-scripts/start-chain/start-chain.sh"
func start_chain(
	instance string,
	start_chain_script string,
	binary string,
	num_nodes uint,
	chain_id string,
	ip_prefix string,
	grpc_port uint,
	json_port uint,
	genesis_mods string,
) {
	cmd := exec.Command("docker", "exec", instance, "/bin/bash",
		start_chain_script, binary, fmt.Sprint(num_nodes), chain_id, ip_prefix, fmt.Sprint(grpc_port), fmt.Sprint(json_port), genesis_mods)

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
		// fmt.Println("out: " + out)
		if out == "done!!!!!!!!" {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func catFileInDocker(instance string, path string) string {
	cmd := exec.Command("docker", "exec", instance, "cat", path)

	bz, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	return string(bz)
	// cmdReader, err := cmd.StdoutPipe()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// cmd.Stderr = cmd.Stdout

	// if err := cmd.Start(); err != nil {
	// 	log.Fatal(err)
	// }

	// scanner := bufio.NewScanner(cmdReader)

	// for scanner.Scan() {
	// 	out := scanner.Text()
	// 	// fmt.Println("out: " + out)
	// 	if out == "done!!!!!!!!" {
	// 		return
	// 	}
	// }
	// if err := scanner.Err(); err != nil {
	// 	log.Fatal(err)
	// }
}
