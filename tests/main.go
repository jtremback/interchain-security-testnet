package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	start_docker("interchain-security-container", "interchain-security-instance", 9090, 26657, 1317, 8545)
	println("docker started?")

	// time.Sleep(1000 * time.Millisecond)

	cmd := exec.Command("docker", "ps")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", stdoutStderr)

	start_chain("interchain-security-instance", "/testnet-scripts/start-chain/start-chain.sh", "interchain-securityd", 3, "provider", "7.7.7", 26657, 9090, ".app_state.gov.voting_params.voting_period = \"60s\"")

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

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Docker puts stuff on StdErr for no good reason so I combine stderr and stdout
	outChannel := make(chan string)

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			outChannel <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			outChannel <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}()

	for {
		out := <-outChannel
		fmt.Println("out: " + out)
		if out == "beacon!!!!!!!!!!" {
			return
		}
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
	// /bin/bash "$DIR/start-chain/start-chain.sh" interchain-securityd 3 provider 7.7.7 26657 9090 '.app_state.gov.voting_params.voting_period = "60s"'
	// docker exec -ti my_container
	cmd := exec.Command("docker", "exec", instance, "/bin/bash",
		start_chain_script, binary, fmt.Sprint(num_nodes), chain_id, ip_prefix, fmt.Sprint(grpc_port), fmt.Sprint(json_port), genesis_mods)

	stdOutChannel := make(chan string)
	stdErrChannel := make(chan string)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	scanStdOut := func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			stdOutChannel <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}

	scanStdErr := func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			stdErrChannel <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}

	go scanStdOut()
	go scanStdErr()

	for {
		select {
		case out := <-stdOutChannel:
			fmt.Println("out: " + out)
			// if out == "the string" {
			// 	return
			// }

		case err := <-stdErrChannel:
			fmt.Println("err: " + err)
		}
	}

}
