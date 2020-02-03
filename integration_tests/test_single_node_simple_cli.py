import time

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

    transfer_amount = 1000000000000
    fee_for_transfer = 100000000
    gas_for_transfer = 25000000

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
        print("Clooect gentxs")
        cmd.collect_gentxs()
        print("Validate genesis")
        cmd.validate_genesis()

        print("Setup class done.")


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
        print("Start test00_get_balance")

        res = cmd.get_balance("wallet", "anna")
        print("Output: ", res)
        assert(int(res["value"]) == self.basic_coin)

        res = cmd.get_balance("wallet", "elsa")
        assert(int(res["value"]) == self.basic_coin)
        print("Done test00_get_balance")


    def test01_transfer_to(self):
        print("Start test01_transfer_to")

        print("Transfer token from elsa to anna")
        cmd.transfer_to(self.wallet_password, self.info_anna['address'], self.transfer_amount,
                        self.fee_for_transfer, self.gas_for_transfer,
                        'address', self.info_elsa['address'])
        print("Tx sent. Waiting for validation")

        time.sleep(10)

        print("Balance checking after transfer..")
        res = cmd.get_balance("wallet", "anna")
        assert(int(res["value"]) == self.basic_coin + self.transfer_amount)

        res = cmd.get_balance("wallet", "elsa")
        assert(int(res["value"]) < self.basic_coin - self.transfer_amount)

        print("Done test01_transfer_to")
