# starlight-demo-impl


### Basic operations:
  - **setup**: create initial stellar's accounts according to starlight spec + additional `htlc_resolution_account` if necessary
  - **open_channel**: open channel according to starlight spec
  - **unilateral_close**: force close channel(publish latest ratchet's tx and settle_with_*'s tx)
  - **simple_payment**: off-chain payment according to starlight spec
  - **htlc_payment**: payment over htlc(details of htlc implementation see below)
    - **htlc_timeout_payment**: receiver of payment does not know **rpreimage** funds back to the sender in **timelock** time
    - **htlc_success_payment**: receiver of payment knows **rpreimage** funds belong to receiver immediately
  - **unilateral_close_with_active_htlc**: like unilateral_close but with active htlc
  - **htlc_resolution_process**

### Usage
repository contains four demos(see help for details)
1) `./starlight-demo-impl -unilateral_close`
  - setup
  - open_channel
  - unilateral_close

2) `./starlight-demo-impl -payment`
  - setup
  - open_channel
  - simple_payment
  - unilateral_close
  
3) `./starlight-demo-impl -htlc_timeout_payment`
  - setup
  - open_channel
  - htlc_timeout_payment
  - unilateral_close_with_active_htlc
  - htlc_resolution_process(htlc timeout)

4) `./starlight-demo-impl -htlc_success_payment`
  - setup
  - open_channel
  - htlc_success_payment
  - unilateral_close_with_active_htlc
  - htlc_resolution_process(htlc success)

### HTLC
original htlc scheme(on Bitcoin blockchain):
https://github.com/bitcoin/bips/blob/master/bip-0199.mediawiki
