import time

import pytest

from .lib import cmd
from .lib.errors import DeadDaemonException

class TestClass():
    proc_ee = None
    proc_friday = None

    chain_id = "testchain"
    moniker = "testnode"

    wallet_elsa = "elsa"
    wallet_anna = "anna"
    wallet_password = "!friday1234@"

    info_elsa = None
    info_anna = None

    basic_coin = 5000000000000
    basic_stake = 100000000

    basic_bond = 10000000000
    bonding_fee = 1000000000
    bonding_gas = 50000000

    transfer_amount = 1000000000000
    transfer_fee = 100000000
    transfer_gas = 25000000

    tx_confirm_delay = 10


    def daemon_healthcheck(self):
        is_ee_alive = cmd.daemon_check(self.proc_ee)
        is_friday_alive = cmd.daemon_check(self.proc_friday)
        if not (is_ee_alive and is_friday_alive):
            if not is_ee_alive:
                print("EE dead")

            if not is_friday_alive:
                print("Friday dead")

            raise DeadDaemonException


    def setup_class(self):
        """
        Make genesis.json and keys
        """
        print("*********************Test class preparation*********************")

        print("Cleanup double check")
        cmd.whole_cleanup()

        print("Init chain")
        cmd.init_chain(self.moniker, self.chain_id)
        print("Copy manifest file")
        cmd.copy_manifest()

        print("Create wallet")
        self.info_elsa = cmd.create_wallet(self.wallet_elsa, self.wallet_password)
        self.info_anna = cmd.create_wallet(self.wallet_anna, self.wallet_password)

        print("Add genesis account in cosmos way")
        cmd.add_genesis_account(self.info_elsa['address'], self.basic_coin, self.basic_stake)
        cmd.add_genesis_account(self.info_anna['address'], self.basic_coin, self.basic_stake)

        print("Add genesis account in EE way")
        cmd.add_el_genesis_account(self.wallet_elsa, self.basic_coin, self.basic_stake)
        cmd.add_el_genesis_account(self.wallet_anna, self.basic_coin, self.basic_stake)

        print("Load chainspec")
        cmd.load_chainspec()

        print("Apply general clif config")
        cmd.clif_configs(self.chain_id)

        print("Gentx")
        cmd.gentx(self.wallet_elsa, self.wallet_password)
        print("Collect gentxs")
        cmd.collect_gentxs()
        print("Validate genesis")
        cmd.validate_genesis()

        print("*********************Setup class done*********************")


    def teardown_class(self):
        """
        Delete all data and configs
        """
        print("Test finished and teardowning")
        cmd.delete_wallet(self.wallet_anna, self.wallet_password)
        cmd.delete_wallet(self.wallet_elsa, self.wallet_password)
        cmd.whole_cleanup()


    def setup_method(self):
        print("Running CasperLabs EE..")
        self.proc_ee = cmd.run_casperlabsEE()
        print("Running friday node..")
        self.proc_friday = cmd.run_node()

        # Waiting for nodef process is up and ready for receiving tx...
        time.sleep(10)

        self.daemon_healthcheck()
        print("Runup done. start testing")


    def teardown_method(self):
        print("Terminating daemons..")
        self.proc_friday.terminate()
        self.proc_ee.terminate()

        print("Reset blocks")
        cmd.unsafe_reset_all()


    def test00_get_balance(self):
        print("======================Start test00_get_balance======================")

        res = cmd.get_balance("wallet", "anna")
        print("Output: ", res)
        assert(int(res["value"]) == self.basic_coin)

        res = cmd.get_balance("wallet", "elsa")
        assert(int(res["value"]) == self.basic_coin)
        print("======================Done test00_get_balance======================")


    def test01_transfer_to(self):
        print("======================Start test01_transfer_to======================")

        print("Transfer token from elsa to anna")
        tx_hash = cmd.transfer_to(self.wallet_password, self.info_anna['address'], self.transfer_amount,
                        self.transfer_fee, self.transfer_gas,
                        'address', self.info_elsa['address'])
        
        print("Tx sent. Waiting for validation")
        time.sleep(self.tx_confirm_delay)

        print("Check whether tx is ok or not")
        is_ok = cmd.is_tx_ok(tx_hash)
        assert(is_ok == True)

        print("Balance checking after transfer..")
        res = cmd.get_balance("wallet", "anna")
        assert(int(res["value"]) == self.basic_coin + self.transfer_amount)

        res = cmd.get_balance("wallet", "elsa")
        assert(int(res["value"]) < self.basic_coin - self.transfer_amount)

        print("======================Done test01_transfer_to======================")


    @pytest.mark.skip(reason="Bond/Unbond tx itself works, but not effective now")
    def test02_bond_and_unbond(self):
        print("======================Start test02_bond_and_unbond======================")

        print("Bonding token")
        bond_tx_hash = cmd.bond(self.wallet_password, self.basic_bond, self.bonding_fee, self.bonding_gas, "wallet", "anna")
        print("Tx sent. Waiting for validation")

        time.sleep(self.tx_confirm_delay)

        print("Check whether tx is ok or not")
        is_ok = cmd.is_tx_ok(bond_tx_hash)
        assert(is_ok == True)

        print("Balance checking after bonding")
        res_before = cmd.get_balance("wallet", "anna")

        print("Try to send more money than bonding. Invalid tx expected")
        tx_hash_after_bond = cmd.transfer_to(self.wallet_password, self.info_elsa['address'], int(int(res_before['value']) - self.basic_bond / 2),
                                             self.transfer_fee, self.transfer_gas, "wallet", "anna")
        
        print("Tx sent. Waiting for validation")
        time.sleep(self.tx_confirm_delay)
        
        print("Check whether tx is ok or not")
        is_ok = cmd.is_tx_ok(tx_hash_after_bond)
        #assert(is_ok == False)

        print("Balance checking after bonding")
        res_after = cmd.get_balance("wallet", "anna")
        # Reason: Just enough value to ensure that tx become invalid
        assert(self.basic_bond < int(res_after["value"]))

        print("Unbond and try to transfer")
        print("Unbond first")
        tx_hash_unbond = cmd.unbond(self.wallet_password, self.basic_bond, self.bonding_fee, self.bonding_gas, "wallet", "anna")
        print("Tx sent. Waiting for validation")

        time.sleep(self.tx_confirm_delay)

        print("Check whether tx is ok or not")
        is_ok = cmd.is_tx_ok(tx_hash_unbond)
        assert(is_ok == True)

        print("Try to transfer. Will be confirmed in this time")
        tx_hash_after_unbond = cmd.transfer_to(self.wallet_password, self.info_elsa['address'], int(int(res_before['value']) - self.basic_bond / 2),
                                               self.transfer_fee, self.transfer_gas, "wallet", "anna")
        
        print("Tx sent. Waiting for validation")
        time.sleep(self.tx_confirm_delay)

        print("Check whether tx is ok or not")
        is_ok = cmd.is_tx_ok(tx_hash_after_unbond)
        assert(is_ok == True)

        print("Balance checking after bonding")
        res_after_after = cmd.get_balance("wallet", "elsa")
        # Reason: Just enough value to ensure that tx become invalid
        assert(int(res_after_after["value"]) == self.basic_coin + int(res_before['value']) - self.basic_bond)

        print("======================Done test02_bond_and_unbond======================")


    @pytest.mark.skip(reason="Stable local, will unmark skip and check the err message after master branch merge")
    def test02_1_simple_bond_unbond(self):
        print("======================Start test02_1_simple_bond_unbond======================")

        print("Bonding token")
        bond_tx_hash = cmd.bond(self.wallet_password, self.basic_bond, self.bonding_fee, self.bonding_gas, "wallet", "anna")
        print("Tx sent. Waiting for validation")

        time.sleep(self.tx_confirm_delay)

        print("Check whether tx is ok or not")
        is_ok = cmd.is_tx_ok(bond_tx_hash)
        assert(is_ok == True)

        print("Unbonding token")
        tx_hash_unbond = cmd.unbond(self.wallet_password, self.basic_bond, self.bonding_fee, self.bonding_gas, "wallet", "anna")
        print("Tx sent. Waiting for validation")

        time.sleep(self.tx_confirm_delay)

        print("Check whether tx is ok or not")
        is_ok = cmd.is_tx_ok(tx_hash_unbond)
        assert(is_ok == True)

        print("======================Done test02_1_simple_bond_unbond======================")


    def _register_nickname(self):
        print("Set nickname")
        tx_hash_nickname = cmd.set_nickname_by_address(self.wallet_password, "anna", self.info_anna['address'])

        print("Tx sent. Waiting for validation")
        time.sleep(self.tx_confirm_delay)

        print("Check whether the Tx is OK or not")
        is_ok = cmd.is_tx_ok(tx_hash_nickname)
        assert(is_ok == True)

        print("Get registered address and compare to the info from the wallet")
        res_info = cmd.get_address("anna")
        assert(res_info['address'] == self.info_anna['address'])

        print("Well registered!")


    def test03_simple_register_nickname(self):
        print("======================Start test03_simple_register_nickname======================")
        self._register_nickname()
        print("======================Done test03_simple_register_nickname======================")


    def test04_transfer_to_by_nickname(self):
        print("======================Start test04_transfer_to_by_nickname======================")

        self._register_nickname()

        print("Try to transfer to nickname recipient")
        tx_hash_transfer = cmd.transfer_to(self.wallet_password, "anna", self.transfer_amount,
                                           self.transfer_fee, self.transfer_gas, "wallet", "elsa")

        print("Tx sent. Waiting for validation")
        time.sleep(self.tx_confirm_delay)

        print("Check Tx OK or not")
        is_ok = cmd.is_tx_ok(tx_hash_transfer)
        assert(is_ok == True)

        print("Check wallet by address. Should be match with wallet info")
        res_transfer = cmd.get_balance("address", self.info_anna['address'])
        assert(int(res_transfer['value']) == int(self.basic_coin + self.transfer_amount))

        print("======================Done test04_transfer_to_by_nickname======================")
        