package main

// property erupt day common remind oblige chunk thumb jazz camera erupt reward divorce fit toy cargo traffic scrub begin gown recall video friend prosper
// decide praise business actor peasant farm drastic weather extend front hurt later song give verb rhythm worry fun pond reform school tumble august one
// brown include source lesson joy fringe great hazard breeze essay hurdle gadget make prepare unfair sense divorce emotion double elite more subway hat worth
// sight similar better jar bitter laptop solve fashion father jelly scissors chest uniform play unhappy convince silly clump another conduct behave reunion marble animal
// glass trip produce surprise diamond spin excess gaze wash drum human solve dress minor artefact canoe hard ivory orange dinner hybrid moral potato jewel
// pave immune ethics wrap gain ceiling always holiday employ earth tumble real ice engage false unable carbon equal fresh sick tattoo nature pupil nuclear

// Attributes that are unique to a validator. Allows us to map (part of) the set of uints to
// a set of viable validators
type ValidatorAttrs struct {
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
	validatorAttrs    []ValidatorAttrs
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
		validatorAttrs: []ValidatorAttrs{
			{
				mnemonic:   "pave immune ethics wrap gain ceiling always holiday employ earth tumble real ice engage false unable carbon equal fresh sick tattoo nature pupil nuclear",
				delAddress: "cosmos19hz4m226ztankqramvt4a7t0shejv4dc79gp9u",
			},
			{
				mnemonic:   "glass trip produce surprise diamond spin excess gaze wash drum human solve dress minor artefact canoe hard ivory orange dinner hybrid moral potato jewel",
				delAddress: "cosmos19pe9pg5dv9k5fzgzmsrgnw9rl9asf7ddwhu7lm",
			},
			{
				mnemonic:   "sight similar better jar bitter laptop solve fashion father jelly scissors chest uniform play unhappy convince silly clump another conduct behave reunion marble animal",
				delAddress: "cosmos1dkas8mu4kyhl5jrh4nzvm65qz588hy9qcz08la",
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
