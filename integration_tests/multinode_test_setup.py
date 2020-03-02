from lib import cmd


def setup_multinode():
    chain_id = "ci_testnet"

    # 'bryan' will be used in account creation & broadcasting in transfer
    wallet_elsa = "elsa"
    wallet_anna = "anna"
    wallet_olaf = "olaf"
    wallet_hans = "hans"
    wallet_bryan = "bryan"
    
    wallet_password = "!friday1234@"

    info_elsa = None
    info_anna = None
    info_olaf = None
    info_hans = None

    basic_coin = 500000000000000000000
    basic_stake = 100000000

    print("*********************Multinode environment preparation start*********************")

    print("Copy manifest file")
    cmd.copy_manifest()

    # 'bryan' are not in genesis account, but create wallets in advance
    print("Create wallet")
    info_elsa = cmd.create_wallet(wallet_elsa, wallet_password)
    info_anna = cmd.create_wallet(wallet_anna, wallet_password)
    info_olaf = cmd.create_wallet(wallet_olaf, wallet_password)
    info_hans = cmd.create_wallet(wallet_hans, wallet_password)
    info_bryan = cmd.create_wallet(wallet_bryan, wallet_password)

    print("Add genesis account in cosmos way")
    cmd.add_genesis_account(info_elsa['address'], basic_coin, basic_stake)
    cmd.add_genesis_account(info_anna['address'], basic_coin, basic_stake)
    cmd.add_genesis_account(info_olaf['address'], basic_coin, basic_stake)
    cmd.add_genesis_account(info_hans['address'], basic_coin, basic_stake)

    print("Add genesis account in EE way")
    cmd.add_el_genesis_account(wallet_elsa, basic_coin, basic_stake)
    cmd.add_el_genesis_account(wallet_anna, basic_coin, basic_stake)
    cmd.add_el_genesis_account(wallet_olaf, basic_coin, basic_stake)
    cmd.add_el_genesis_account(wallet_hans, basic_coin, basic_stake)

    print("Load chainspec")
    cmd.load_chainspec()

    print("Apply general clif config")
    cmd.clif_configs(chain_id)

    print("Gentx")
    cmd.gentx(wallet_elsa, wallet_password)
    print("Collect gentxs")
    cmd.collect_gentxs()
    print("Validate genesis")
    cmd.validate_genesis()

    print("*********************Multinode environment preparation done*********************")


if __name__ == "__main__":
    setup_multinode()
    
