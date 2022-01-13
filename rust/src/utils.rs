// THIS FILE WAS COPY & PASTED from the cosmos-gravity-bridge repo because it appeared in a
// binary crate which was not possible to import from.

use clarity::Address as EthAddress;
use clarity::PrivateKey as EthPrivateKey;
use deep_space::private_key::PrivateKey as CosmosPrivateKey;
use deep_space::Contact;
use std::process::Command;
use std::{fs::File, path::Path};
use std::{
    io::{BufRead, BufReader, Read, Write},
    process::ExitStatus,
};

#[derive(Debug, Clone)]
pub struct ValidatorKeys {
    /// The Ethereum key used by this validator to sign Gravity bridge messages
    pub eth_key: EthPrivateKey,
    /// The Orchestrator key used by this validator to submit oracle messages and signatures
    /// to the cosmos chain
    pub orch_key: CosmosPrivateKey,
    /// The validator key used by this validator to actually sign and produce blocks
    pub validator_key: CosmosPrivateKey,
}

/// Ethereum private keys for the validators are generated using the gravity eth_keys add command
/// and dumped into a file /validator-eth-keys in the container, from there they are then used by
/// the orchestrator on startup
pub fn parse_ethereum_keys() -> Vec<EthPrivateKey> {
    let filename = "../docker/validator-eth-keys";
    let file = File::open(filename).expect("Failed to find eth keys");
    let reader = BufReader::new(file);
    let mut ret = Vec::new();

    for line in reader.lines() {
        let key = line.expect("Error reading eth key file!");
        if key.is_empty() || key.contains("public") || key.contains("address") {
            continue;
        }
        let key = key.split(':').last().unwrap().trim();
        ret.push(key.parse().unwrap());
    }
    ret
}

/// Parses the output of the cosmoscli keys add command to import the private key
fn parse_phrases(filename: &str) -> Vec<CosmosPrivateKey> {
    let file = File::open(filename).expect("Failed to find phrases");
    let reader = BufReader::new(file);
    let mut ret = Vec::new();

    for line in reader.lines() {
        let phrase = line.expect("Error reading phrase file!");
        if phrase.is_empty()
            || phrase.contains("write this mnemonic phrase")
            || phrase.contains("recover your account if")
        {
            continue;
        }
        let key = CosmosPrivateKey::from_phrase(&phrase, "").expect("Bad phrase!");
        ret.push(key);
    }
    ret
}

/// Validator private keys are generated via the gravity key add
/// command, from there they are used to create gentx's and start the
/// chain, these keys change every time the container is restarted.
/// The mnemonic phrases are dumped into a text file /validator-phrases
/// the phrases are in increasing order, so validator 1 is the first key
/// and so on. While validators may later fail to start it is guaranteed
/// that we have one key for each validator in this file.
pub fn parse_validator_keys() -> Vec<CosmosPrivateKey> {
    let filename = "../docker/validator-phrases";
    parse_phrases(filename)
}

/// Orchestrator private keys are generated via the gravity key add
/// command just like the validator keys themselves and stored in a
/// similar file /orchestrator-phrases
pub fn parse_orchestrator_keys() -> Vec<CosmosPrivateKey> {
    let filename = "../docker/orchestrator-phrases";
    parse_phrases(filename)
}

pub fn get_keys() -> Vec<ValidatorKeys> {
    let cosmos_keys = parse_validator_keys();
    let orch_keys = parse_orchestrator_keys();
    let eth_keys = parse_ethereum_keys();
    let mut ret = Vec::new();
    for ((c_key, o_key), e_key) in cosmos_keys.into_iter().zip(orch_keys).zip(eth_keys) {
        ret.push(ValidatorKeys {
            eth_key: e_key,
            validator_key: c_key,
            orch_key: o_key,
        })
    }
    ret
}
