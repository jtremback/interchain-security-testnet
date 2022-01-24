package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

// property erupt day common remind oblige chunk thumb jazz camera erupt reward divorce fit toy cargo traffic scrub begin gown recall video friend prosper
// decide praise business actor peasant farm drastic weather extend front hurt later song give verb rhythm worry fun pond reform school tumble august one
// brown include source lesson joy fringe great hazard breeze essay hurdle gadget make prepare unfair sense divorce emotion double elite more subway hat worth
// sight similar better jar bitter laptop solve fashion father jelly scissors chest uniform play unhappy convince silly clump another conduct behave reunion marble animal
// glass trip produce surprise diamond spin excess gaze wash drum human solve dress minor artefact canoe hard ivory orange dinner hybrid moral potato jewel
// pave immune ethics wrap gain ceiling always holiday employ earth tumble real ice engage false unable carbon equal fresh sick tattoo nature pupil nuclear

func main() {
	start_docker("interchain-security-container", "interchain-security-instance", 9090, 26657, 1317, 8545)
	println("docker started?")

	start_chain("interchain-security-instance", "/testnet-scripts/start-chain/start-chain.sh", "interchain-securityd",
		`[
			"pave immune ethics wrap gain ceiling always holiday employ earth tumble real ice engage false unable carbon equal fresh sick tattoo nature pupil nuclear", 
			"glass trip produce surprise diamond spin excess gaze wash drum human solve dress minor artefact canoe hard ivory orange dinner hybrid moral potato jewel", 
			"sight similar better jar bitter laptop solve fashion father jelly scissors chest uniform play unhappy convince silly clump another conduct behave reunion marble animal"
		]`,
		"provider", "7.7.7", 26657, 9090, ".app_state.gov.voting_params.voting_period = \"60s\"")

	println("chain started?")

	// ms := 0
	// for {
	// 	println(fmt.Sprint(ms) + " MILLISECONDS ----------------------------------------------")

	// 	bz, _ := exec.Command("docker", "exec", "interchain-security-instance", "interchain-securityd", "query", "bank", "balances",
	// 		`cosmos19pe9pg5dv9k5fzgzmsrgnw9rl9asf7ddwhu7lm`,
	// 		`--chain-id`, `provider`,
	// 		`--home`, `/provider/validator1`,
	// 	).CombinedOutput()
	// 	fmt.Println(string(bz))

	// 	// docker exec interchain-security-instance interchain-securityd query block
	// 	bz, _ = exec.Command("docker", "exec", "interchain-security-instance", "interchain-securityd", "query", "block",
	// 		`--chain-id`, `provider`,
	// 		`--home`, `/provider/validator1`,
	// 	).CombinedOutput()
	// 	// fmt.Println(string(bz) == `{"block_id":{"hash":"","parts":{"total":0,"hash":""}},"block":null}`)
	// 	fmt.Println(regexp.Match(`{"block_id":{"hash":"","parts":{"total":0,"hash":""}},"block":null}`, bz))

	// 	time.Sleep(100 * time.Millisecond)
	// 	ms += 100
	// }

	// // docker exec interchain-security-instance interchain-securityd tx gov submit-proposal --title="Test Proposal" --description="My awesome proposal" --type Text --deposit 10000000stake --from validator1 --chain-id provider --home /provider/validator1 --keyring-backend test
	// bz, _ := exec.Command("docker", "exec", "interchain-security-instance", "interchain-securityd", "tx", "gov", "submit-proposal",
	// 	`--title="Test Proposal"`,
	// 	`--description="My awesome proposal"`,
	// 	`--type`, `Text`,
	// 	`--deposit`, `10000000stake`,
	// 	`--from`, `validator1`,
	// 	`--chain-id`, `provider`,
	// 	`--home`, `/provider/validator1`,
	// 	`--keyring-backend`, `test`,
	// ).CombinedOutput()

	// fmt.Println(string(bz))

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

	outChannel := make(chan string)
	<-outChannel
}

// Waits a number of blocks
func waitBlocks(num uint) {
	for {

		time.Sleep(100 * time.Millisecond)
	}
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
	mnemonics string,
	chain_id string,
	ip_prefix string,
	grpc_port uint,
	json_port uint,
	genesis_mods string,
) {
	cmd := exec.Command("docker", "exec", instance, "/bin/bash",
		start_chain_script, binary, mnemonics, chain_id, ip_prefix, fmt.Sprint(grpc_port), fmt.Sprint(json_port), genesis_mods)

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
		fmt.Println("start chain: " + out)
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
