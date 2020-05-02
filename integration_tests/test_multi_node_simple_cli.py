import time
import random

import pytest

from .lib import cmd
from .lib.errors import DeadDaemonException


class TestMultiNodeSimple:
    network_runup_delay = 20
    nodes_ssh = ["ubuntu@140.238.12.186", "opc@132.145.83.49", "opc@150.136.172.164", "ubuntu@140.238.73.77"]
    nodes_address = [
              "tcp://localhost:26657"
            , "tcp://132.145.83.49:26657"
            , "tcp://150.136.172.164:26657"
            , "tcp://140.238.73.77:26657"
        ]

    chain_id = "ci_testnet"
    
    monikers = ["node1", "node2", "node3", "node4"]
    bls_pubkeys = [] # will be filled in setup phase
    
    wallet_elsa = "elsa"
    wallet_anna = "anna"
    wallet_olaf = "olaf"
    wallet_hans = "hans"
    wallet_bryan = "bryan"
    wallet_password = "!friday1234@"

    nickname_anna = "princessanna"
    nickname_elsa = "princesselsa"

    info_elsa = None
    info_anna = None
    info_olaf = None
    info_hans = None
    info_bryan = None

    basic_coin = "1000000000000000000000000000"
    basic_stake = "1000000000000000000"

    multiplier = 10 ** 18

    basic_coin_amount = int(int(basic_coin) / multiplier)

    basic_bond = "1"
    bonding_fee = "0.001"

    transfer_amount = "1"
    transfer_fee = "0.001"

    tx_blocktime = 6


    def get_node_randomly(self):
        return random.choice(self.nodes_address[:3])


    def setup_class(self):
        """
        Register as validator & bonding asset
        """
        print("*********************Test class preparation*********************")

        print("Gather wallet info")
        self.info_elsa = cmd.get_wallet_info(self.wallet_elsa)
        self.info_anna = cmd.get_wallet_info(self.wallet_anna)
        self.info_olaf = cmd.get_wallet_info(self.wallet_olaf)
        self.info_hans = cmd.get_wallet_info(self.wallet_hans)
        self.info_bryan = cmd.get_wallet_info(self.wallet_bryan)

        print("Gather BLS public key (valconspub key)")
        for node in self.nodes_ssh:
            valconspub = cmd.get_bls_pubkey_remote(node)
            self.bls_pubkeys.append(valconspub)

        print("Create validators...")
        # 'bryan' is not validator
        val_tx_hashes = []
        for unit_node_address, unit_pub, unit_wallet_alias, unit_moniker in \
                zip(self.nodes_address[1:], self.bls_pubkeys[1:], [self.wallet_anna, self.wallet_hans, self.wallet_olaf], self.monikers[1:]):

            print("For {}".format(unit_node_address))
            unit_val_tx_hash = cmd.create_validator(self.wallet_password, unit_wallet_alias, unit_pub, unit_moniker)
            val_tx_hashes.append(unit_val_tx_hash)

        print("Wait for switching next block...")
        time.sleep(60)

        print("Bonding for validators..")
        bond_hashes = []
        for unit_node_address, unit_pub, unit_wallet_alias, unit_moniker in \
                zip(self.nodes_address[1:], self.bls_pubkeys[1:], [self.wallet_anna, self.wallet_hans, self.wallet_olaf], self.monikers[1:]):

            print("For {}".format(unit_node_address))
            unit_bond_hash = cmd.bond(self.wallet_password, self.basic_bond, self.bonding_fee, unit_wallet_alias)
            bond_hashes.append(unit_bond_hash)

        print("Wait for switching next block...")
        time.sleep(30)

        cmd._process_executor("curl http://localhost:26657/num_unconfirmed_txs", need_output=True)
        print("Check whether all txs are valid or not..")
        for unit_val_tx, unit_bond_tx in zip(val_tx_hashes, bond_hashes):
            val_tx_ok = cmd.is_tx_ok(unit_val_tx)
            bond_tx_ok = cmd.is_tx_ok(unit_bond_tx)

            assert(val_tx_ok)
            assert(bond_tx_ok)

        print("Wait a while preparing validation...")
        time.sleep(self.tx_blocktime * 3 + 1)

        print("*********************Setup class done*********************")


    def teardown_class(self):
        """
        Cannot cleanup & node restarting in this environment.
        """
        print("Test finished and teardowning")


    def setup_method(self):
        """
        No setup of tests
        """
        pass


    def teardown_method(self):
        """
        No teardown of tests
        """
        pass


    def test00_get_balance(self):
        print("======================Start test00_get_balance======================")

        for node in self.nodes_address[:3]:
            print("Test of {}".format(node))
            res = cmd.get_balance(self.wallet_anna, node=node)
            assert("value" in res)

            res = cmd.get_balance(self.wallet_elsa, node=node)
            assert("value" in res)

            print("{} OK".format(node))

        print("======================Done test00_get_balance======================")


    def test01_00_transfer_to(self):
        print("======================Start test01_transfer_to======================")

        print("Transfer token from elsa to anna using")

        picked_node = self.get_node_randomly()
        print("At {}".format(picked_node))

        tx_hash = cmd.transfer_to(self.wallet_password, self.info_anna['address'], self.transfer_amount,
                        self.transfer_fee, self.info_elsa['address'], node=picked_node)

        print("Tx sent. Waiting for validation")
        time.sleep(self.tx_blocktime * 3 + 1)

        print("Check whether tx is ok or not")
        is_ok = cmd.is_tx_ok(tx_hash)
        assert(is_ok == True)


        print("Balance checking after transfer")

        picked_node = self.get_node_randomly()
        print("At {}".format(picked_node))

        res = cmd.get_balance(self.wallet_anna, node=picked_node)
        assert((self.basic_coin_amount + float(self.transfer_amount)) * 0.95 < float(res))

        picked_node = self.get_node_randomly()
        print("At {}".format(picked_node))
        res = cmd.get_balance(self.wallet_elsa, node=picked_node)
        assert(float(res) < self.basic_coin_amount - float(self.transfer_amount))

        print("======================Done test01_transfer_to======================")

    def test01_01_transfer_to_nonexistent_account(self):
        print("======================Start test01_1_transfer_to_nonexistent_account======================")
        print("Transfer token from anna to bryan")

        picked_node = self.get_node_randomly()
        print("At {}".format(picked_node))

        tx_hash = cmd.transfer_to(self.wallet_password, self.info_bryan['address'], float(self.transfer_amount) / 10,
                        self.transfer_fee, self.info_anna['address'], node=picked_node)

        print("Tx sent. Waiting for validation")
        time.sleep(self.tx_blocktime * 3 + 1)

        print("Check whether tx is ok or not")
        is_ok = cmd.is_tx_ok(tx_hash)
        assert(is_ok == True)

        print("Query balance to each node.")
        # Check the existence of receipient account
        # As this multinode test has no reset phase,
        #   because of multinode controlling issue in python level test code.
        # So, we cannot sure of the status before the test start.
        # 
        # If the query result comes, we can sure that the account exists, although there is no balance assertion.
        for node in self.nodes_address[:3]:
            print("Test of {}".format(node))
            res = cmd.get_balance(self.wallet_anna, node=node)
            print("Balance of 'anna': ", float(res))

            res = cmd.get_balance(self.wallet_bryan, node=node)
            print("Balance of 'anna': ", float(res))
            print("{} OK".format(node))

        print("======================End test01_1_transfer_to_nonexistent_account======================")


    def _register_nickname(self):
        print("Set nickname")
        picked_node = self.get_node_randomly()
        print("At {}".format(picked_node))
        tx_hash_nickname = cmd.set_nickname(self.wallet_password, self.nickname_anna, self.info_anna['address'], node=picked_node)

        print("Tx sent. Waiting for validation")
        time.sleep(self.tx_blocktime * 3 + 1)

        print("Check whether the Tx is OK or not")
        is_ok = cmd.is_tx_ok(tx_hash_nickname)
        assert(is_ok == True)

        print("Get registered address and compare to the info from the wallet")
        picked_node = self.get_node_randomly()
        print("At {}".format(picked_node))

        res_info = cmd.get_address(self.nickname_anna, node=picked_node)
        assert(res_info['address'] == self.info_anna['address'])

        print("Well registered!")


    def test03_simple_register_nickname(self):
        print("======================Start test03_simple_register_nickname======================")
        self._register_nickname()
        print("======================Done test03_simple_register_nickname======================")


    def test04_transfer_to_by_nickname(self):
        print("======================Start test04_transfer_to_by_nickname======================")
        print("Try to transfer to nickname recipient")

        picked_node = self.get_node_randomly()
        print("At {}".format(picked_node))

        tx_hash_transfer = cmd.transfer_to(self.wallet_password, self.nickname_anna, self.transfer_amount,
                                           self.transfer_fee, self.wallet_elsa, node=picked_node)

        print("Tx sent. Waiting for validation")
        time.sleep(self.tx_blocktime * 3 + 1)

        print("Check Tx OK or not")
        is_ok = cmd.is_tx_ok(tx_hash_transfer)
        assert(is_ok == True)

        # We cannot sure the status of each wallet. Just validate Tx and skip value comparison
        #res_transfer = cmd.get_balance("address", self.info_anna['address'], node=picked_node)
        #assert(int(res_transfer) == int(self.basic_coin + self.transfer_amount))

        print("Try to transfer to nickname sender")
        tx_hash_transfer = cmd.transfer_to(self.wallet_password, self.info_elsa['address'], self.transfer_amount,
                                           self.transfer_fee, self.nickname_anna)

        print("Tx sent. Waiting for validation")
        time.sleep(self.tx_blocktime * 3 + 1)

        print("Check Tx OK or not")
        is_ok = cmd.is_tx_ok(tx_hash_transfer)
        assert(is_ok == True)

        print("======================Done test04_transfer_to_by_nickname======================")
