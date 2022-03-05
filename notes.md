	// Pass gov prop
	// Query IBC module for client state and consensus state????? <- NO but should be possible
	
	// Instead, create client state:
	// clientState = ibctmtypes.NewClientState(<provider chain id>, ibctmtypes.DefaultTrustLevel, <unbonding period- query provider staking params>, <half of the previous argument>,
	// 		time.Second*10, <height at which the proposal passed on provider +-1>, commitmenttypes.GetSDKSpecs(), []string{"upgrade", "upgradedIBCState"}, true, true),
	
	// Then, create consensus state:
	// ConsensusState = ibctmtypes.NewConsensusState(<time at which the proposal passed on provider +-1>, commitmenttypes.NewMerkleRoot(<AppHash from block header from when the proposal passed on the provider +-1>), <NextValidatorsHash from block header from when the proposal passed on the provider +-1>)
	// Populate consumer genesis - set enabled to true
	// start consumer chain

	// Set up relayer- Hermes may be best
	// https://hermes.informal.systems/config.html
	// - rpc endpoint
	// - private keys of an address with money
	// - probably mostly change id-key_name in config
	// - Massive headache: off by one errors

	consState = GetSelfConsensusState(ctx.blockheight (+-1???))


	clientState := k.GetTemplateClient(ctx)
	clientState.ChainId = ctx.chainid
	clientState.LatestHeight = ctx.blockheight (+-1???)
	clientState.TrustingPeriod = unbondingTime / 2
	clientState.UnbondingPeriod = unbondingTime

	gen.initialvalset = tm.validatorset -> utility func


Things that are wrong with the 2 chain setup
- Not using "node" argument to query correct node
- Collision on pprof port
- Need persistent peers arg in start command

# magically faster
skip_timeout_commit = true

or 

timeout_commit = "10ms" # or something else short

might work or might not

peer_gossip_sleep_duration = "10ms" or lower

this one is probably good to make faster

flush_throttle_timeout = "10ms" or lower



// property erupt day common remind oblige chunk thumb jazz camera erupt reward divorce fit toy cargo traffic scrub begin gown recall video friend prosper
// decide praise business actor peasant farm drastic weather extend front hurt later song give verb rhythm worry fun pond reform school tumble august one
// brown include source lesson joy fringe great hazard breeze essay hurdle gadget make prepare unfair sense divorce emotion double elite more subway hat worth
// sight similar better jar bitter laptop solve fashion father jelly scissors chest uniform play unhappy convince silly clump another conduct behave reunion marble animal
// glass trip produce surprise diamond spin excess gaze wash drum human solve dress minor artefact canoe hard ivory orange dinner hybrid moral potato jewel
// pave immune ethics wrap gain ceiling always holiday employ earth tumble real ice engage false unable carbon equal fresh sick tattoo nature pupil nuclear

// priv_validator_key
// {"address":"06C0F3E47CC5C748269088DC2F36411D3AAA27C6","pub_key":{"type":"tendermint/PubKeyEd25519","value":"RrclQz9bIhkIy/gfL485g3PYMeiIku4qeo495787X10="},"priv_key":{"type":"tendermint/PrivKeyEd25519","value":"uX+ZpDMg89a6gtqs/+MQpCTSqlkZ0nJQJOhLlCJvwvdGtyVDP1siGQjL+B8vjzmDc9gx6IiS7ip6jj3nvztfXQ=="}}
// node_key
// {"priv_key":{"type":"tendermint/PrivKeyEd25519","value":"fjw4/DAhyRPnwKgXns5SV7QfswRSXMWJpHS7TyULDmJ8ofUc5poQP8dgr8bZRbCV5RV8cPqDq3FPdqwpmUbmdA=="}}

// priv_validator_key
// {"address":"99BD3A72EF12CD024E7584B3AC900AE3743C6ADF","pub_key":{"type":"tendermint/PubKeyEd25519","value":"mAN6RXYxSM4MNGSIriYiS7pHuwAcOHDQAy9/wnlSzOI="},"priv_key":{"type":"tendermint/PrivKeyEd25519","value":"QePcwfWtOavNK7pBJrtoLMzarHKn6iBWfWPFeyV+IdmYA3pFdjFIzgw0ZIiuJiJLuke7ABw4cNADL3/CeVLM4g=="}}
// node_key
// {"priv_key":{"type":"tendermint/PrivKeyEd25519","value":"TQ4vHcO/vKdzGtWpelkX53WdMQd4kTsWGFrdcatdXFvWyO215Rewn5IRP0FszPLWr2DqPzmuH8WvxYGk5aeOXw=="}}

// priv_validator_key
// {"address":"C888306A908A217B9A943D1DAD8790044D0947A4","pub_key":{"type":"tendermint/PubKeyEd25519","value":"IHo4QEikWZfIKmM0X+N+BjKttz8HOzGs2npyjiba3Xk="},"priv_key":{"type":"tendermint/PrivKeyEd25519","value":"z08bmSB91uFVpVmR3t2ewd/bDjZ/AzwQpe5rKjWiPG0gejhASKRZl8gqYzRf434GMq23Pwc7MazaenKOJtrdeQ=="}}
// node_key
// {"priv_key":{"type":"tendermint/PrivKeyEd25519","value":"WLTcHEjbwB24Wp3z5oBSYTvtGQonz/7IQabOFw85BN0UkkyY5HDf38o8oHlFxVI26f+DFVeICuLbe9aXKGnUeg=="}}

// priv_validator_key
// {"address":"BAE65B1FA13E12423FB394FDF6D0A9579B1345D1","pub_key":{"type":"tendermint/PubKeyEd25519","value":"98GmDXdD6ATd+pAeHuITLtbCI8MyRpT3f8viDJzM7W0="},"priv_key":{"type":"tendermint/PrivKeyEd25519","value":"YYM/LDqDaeGVqtlCobRQBZzO0bXi1Fi8hOppeoXcYU73waYNd0PoBN36kB4e4hMu1sIjwzJGlPd/y+IMnMztbQ=="}}
// node_key
// {"priv_key":{"type":"tendermint/PrivKeyEd25519","value":"u1LPPCMX2SFvTXX69Wa7BsxoBAK53dlkqq2frRivDrx3pmxBHn2b+hSn+FXd5T+TyoybigJit0gDJCoq4v0phQ=="}}

// priv_validator_key
// {"address":"20470D5AF05713468DCBC0E487EA3864AC904B78","pub_key":{"type":"tendermint/PubKeyEd25519","value":"y/L3RtCVM/RC6ZLCCQvRDgl5Ol5bXaYVBbzXd2rcAFk="},"priv_key":{"type":"tendermint/PrivKeyEd25519","value":"aX97wKlonMcE5sbYMLSML35s9ZrXhulJYvTV3dYtF+zL8vdG0JUz9ELpksIJC9EOCXk6XltdphUFvNd3atwAWQ=="}}
// node_key
// {"priv_key":{"type":"tendermint/PrivKeyEd25519","value":"AnLJOQpkt9SjmMoh0qEjI2A9sY6uIDPlUR6Qgj7MuOJsIroS65HeV612IIEP3WfQywREfmMez5F9XxaE+1eHXA=="}}

// priv_validator_key
// {"address":"6376991142228EBD242CA253D7612EE7CDB12C53","pub_key":{"type":"tendermint/PubKeyEd25519","value":"/EFyxiOz1zuFEG0vF4FAj9NZSaa0Iv90pb23JBCz9nA="},"priv_key":{"type":"tendermint/PrivKeyEd25519","value":"GkS4m705sfiu1j1pxiyZYsYOngxR72mBRbyaBfwFVNf8QXLGI7PXO4UQbS8XgUCP01lJprQi/3SlvbckELP2cA=="}}
// node_key
// {"priv_key":{"type":"tendermint/PrivKeyEd25519","value":"lSJtiqeO9OjaCXwugD3QOfO2wHGQgm2vjnLt0B1lTYrpUhlpDmc9DbikElSp1Lswv4wjG/4ZgXa/CKbP7JtPDg=="}}

// NUM=2; echo priv_validator_key; cat provider/validator$NUM/config/priv_validator_key.json | jq -c; echo node_key; cat provider/validator$NUM/config/node_key.json | jq -c


