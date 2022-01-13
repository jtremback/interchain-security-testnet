--------------------------- MODULE GravityMBT ---------------------------

(*
This spec is intended to model the behaviors of several different actors interacting with a gravity bridge,
in order to test the Gravity Cosmos Module:

- Cosmos and Ethereum users.
- The Gravity.sol Solidity contract and the Ethereum blockchain.
- "Orchestrator" binaries run by the validators of the Cosmos chain.
*)

\* This early version just sends cosmos coins onto Ethereum.

EXTENDS Integers

CONSTANTS
    \* @type: Int -> Int
    validators

VARIABLES
    (* @type: [
        actionType: Str,
        validator: Int,
        sendAmount: Int,
        eventNonce: Int,
    ]*)
    action,
    \* @type: Bool
    erc20Deployed,
    \* @type: Int
    eventNonce

Init ==
    /\  action = [ actionType |-> "Init" ]
    /\  erc20Deployed = FALSE
    /\  eventNonce = 1

SendToEthereum ==
    /\  erc20Deployed = TRUE
    /\  \E v \in validators:
            action' = [ actionType |-> "SendToEthereum", validator |-> v, sendAmount |-> 1 ]
    /\  UNCHANGED <<erc20Deployed, eventNonce>>

DeployERC20 ==
    /\  eventNonce' = eventNonce + 1
    /\  erc20Deployed' = TRUE
    /\  action' = [  actionType |-> "Erc20DeployedEvent", eventNonce |-> eventNonce ]

Next ==
    \/  SendToEthereum
    \/  DeployERC20
===============================================================================