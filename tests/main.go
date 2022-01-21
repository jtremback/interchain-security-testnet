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

	start_chain("interchain-security-instance", "/testnet-scripts/start-chain/start-chain.sh", "interchain-securityd", 3, "provider", "7.7.7", 26657, 9090, ".app_state.gov.voting_params.voting_period = \"60s\"")
	println("chain started?")

	mnemonic1 := strings.Split(catFileInDocker("interchain-security-instance", "/provider/validator1/mnemonic"), "\n")[5]

	println(mnemonic1)

	bz, _ := exec.Command("interchain-securityd", "keys", "delete", "validator1", "-y").CombinedOutput()
	fmt.Println(string(bz))

	cmd := exec.Command("interchain-securityd", "keys", "add", "validator1", "--recover")
	cmd.Stdin = strings.NewReader(mnemonic1 + "\npassword\npassword")
	bz, err := cmd.CombinedOutput()

	fmt.Println(string(bz))

	if err != nil {
		log.Fatal(err)
	}

	bz, _ = exec.Command("interchain-securityd", "keys", "list").CombinedOutput()

	fmt.Println(string(bz))

	// interchain-securityd tx gov submit-proposal --title="Test Proposal" --description="My awesome proposal" --type Text --deposit 10000000stake --from validator1 --dry-run
	bz, _ = exec.Command("interchain-securityd", "tx", "gov", "submit-proposal", `--title="Test Proposal"`, `--description="My awesome proposal"`, `--type`, `Text`, `--deposit`, `10000000stake`, `--from`, `validator1`, `--chain-id`, `provider`).CombinedOutput()

	fmt.Println(string(bz))

	outChannel := make(chan string)
	<-outChannel
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
		fmt.Println("out: " + out)
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
