# Testing package

## Package structure

```plain
integration_test
  |- jenkins_jobs: Bash backup of Jenkins jobs
  |- lib
    |- cmd.py                           : CLI wrapper
    |- error.py                         : Error suite
  |- test_single_node_simple_cli.py     : Single node CLI kind-of CLI unit test. Runs in Travis CI

  |- config_setting.py                  : Used in multinode test. Making `config.toml`
  |- multinode_test_setup.py            : Create wallet, set genesis account, and signing genesis tx
  |- test_multi_node_simple_cli.py      : Multi node CLI test for checking broadcasting. Runs in Jenkins
```

## Prerequisite

* Python >= 3.6
* `cd integration_test && pip3 install -r ./requirements.txt`

## How to run

* Single node test
`pytest -s test_single_node_simple_cli.py`
* Multi node test
  * Ask @psy2848048 to make ID of [Jenkins](http://132.145.81.228:8080/)
  * Go to `00_CI multinode test`
  * Put your branch

## When & How to put your test case

* New CLI command is created
* After bug fix, the fix & its reproduce test should be included
* Input command wrapper in `lib/cmd.py`
* Write related test case to `test_single_node_simple_cli.py` & `test_multi_node_simple_cli.py`
* The name of the test method should be started from `test`. `pytest` framework gather tests by the name of the methods

## Current test case

1. Single node
    * `00_get_balance`: Get balance by wallet alias
    * `01_transfer_to`
        1. Transfer another account
        1. Check balance both of them
    * `test02_bond_and_unbond`:
        1. Bond token and check whether the tx is valid or not
        1. Check balance whether the amount of transferable tokens is (initial amount - bonded amount)
        1. Try to transfer, less than whole amount, but more than the rest of transferable amount. Expected failure
        1. Unbond tiny amount
        1. Try to transfer, less than whole amount, but more than the rest of transferable amount. Expected success
    * `test03_simple_register_nickname`
        1. Register nickname
        1. Get address of the nickname and compare to local wallet query
    * `test04_transfer_to_by_nickname`
        1. Set nickname
        1. Transfer by nickname recipient
        1. Check balance by nickname
    * `test05_custom_contract_execution`
        1. Try to execute bonding contract by WASM execution
        1. Parameter are custom JSON
        1. Check whether executed success or not
    * `test06_simple_delegate_and_undelegate`
        1. Delegate token and check delegation status
        1. Redelegate token and check delegation status
        1. Undelegate token and check delegation status
    * `test07_simple_vote_and_unvote`
        1. Vote to URef contract
        1. Vote to Hash contract
        1. Unvote to Hash contract
    * `test08_simple_claim_reward_and_commission`
        1. Sleep for inflation proceeding
        1. Get commission & reward rating
        1. Claim reward token
        1. Claim commission token
        1. Check balance
    * `test09_fail_to_tx_lack_of_gas`
        1. Try to transfer with small amount of gas. Should fail
1. Multi node
    * `setup_class`: `create_validator` & `bond` some token
    * `00_get_balance`: Get balance by wallet alias
    * `01_transfer_to`
        1. Pick random node, and send token to exist account
        1. Pick random node, and check balance by wallet alias
    * `test01_01_transfer_to_nonexistent_account`
        1. Pick random node, and send token to non-exist account
        1. Pick random node, and check balance by wallet alias (Checking account broadcast)
    * `test03_simple_register_nickname`
        1. Pick random node, and register nickname
        1. Pick random node, and get address of the nickname and compare to local wallet query
    * `test04_transfer_to_by_nickname`
        1. Pick random node, and transfer token from one account by wallet alias, to another account by nickname
        1. Pick random node, and transfer token from one account by nickname, to another account by address
