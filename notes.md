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



