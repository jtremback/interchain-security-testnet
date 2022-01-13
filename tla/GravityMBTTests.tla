------------------------------ MODULE GravityMBTTests --------------------------------

EXTENDS GravityMBT

SentToEthereumTest ==
    /\  erc20Deployed = TRUE
    /\  action.actionType = "SendToEthereum"

Neg == ~SentToEthereumTest

===============================================================================