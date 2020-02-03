import os
import os.path
import subprocess
import shlex
import json
import re
import shutil
import time

import pexpect


from .errors import DeadDaemonException, FinishedWithError

def _process_executor(cmd: str, *args, need_output=False):
    child = pexpect.spawn(cmd.format(*args))    
    outs = child.read().decode('utf-8')

    if need_output == True:
        res = json.loads(outs)
        return res


def _tx_executor(cmd: str, passphrase, *args):
    try:
        child = pexpect.spawn(cmd.format(*args))
        _ = child.read_nonblocking(10000, timeout=1)
        _ = child.sendline('Y')
        _ = child.read_nonblocking(10000, timeout=1)
        _ = child.sendline(passphrase)
        
        print(child.read())

    except pexpect.TIMEOUT as err:
        raise FinishedWithError


#################
## Daemon control
#################

def run_casperlabsEE(ee_bin="../CasperLabs/execution-engine/target/release/casperlabs-engine-grpc-server",
                     socket_path=".casperlabs/.casper-node.sock") -> subprocess.Popen:
    """
    ./casperlabs-engine-grpc-server $HOME/.casperlabs/.casper-node.sock
    """
    cmd = "{} {}".format(ee_bin, os.path.join(os.environ['HOME'], socket_path))
    proc = subprocess.Popen(shlex.split(cmd), stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    return proc


def run_node() -> subprocess.Popen:
    """
    nodef start
    """
    proc = subprocess.Popen(shlex.split("nodef start"), stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    return proc


def daemon_check(proc: subprocess.Popen):
    """
    Get proc object and check whether given daemon is running or not
    """
    is_alive = proc.poll() is None
    return is_alive



#################
## Setup CLI
#################

def init_chain(moniker: str, chain_id: str) -> subprocess.Popen:
    """
    nodef init <moniker> --chain-id <chain-id>
    """
    _ = _process_executor("nodef init {} --chain-id {}", moniker, chain_id)


def copy_manifest():
    path = os.path.join(os.environ["HOME"], ".nodef/config")
    cmd = "cp ../x/executionlayer/resources/manifest.toml {}".format(path)
    _ = _process_executor(cmd, need_output=False)


def create_wallet(wallet_alias: str, passphrase: str):
    """
    clif key add <wallet_alias>
    """
    try:
        child = pexpect.spawn("clif keys add {}".format(wallet_alias))
        _ = child.read_nonblocking(10000, timeout=1)
        _ = child.sendline(passphrase)
        _ = child.read_nonblocking(10000, timeout=1)
        _ = child.sendline(passphrase)
        
        outs = child.read().decode('utf-8')

    except pexpect.TIMEOUT:
        raise FinishedWithError

    address = None
    pubkey = None
    mnemonic = None

    try:
        # If output keeps json form
        res = json.loads(outs)
        address = res['address']
        pubkey = res['pubkey']
        mnemonic = res['mnemonic']

    except json.JSONDecodeError:
        # If output is not json
        address = re.search(r"address: ([a-z0-9]+)", outs).group(1)
        pubkey = re.search(r"pubkey: ([a-z0-9]+)", outs).group(1)
        mnemonic = outs.strip().split('\n')[-1]

    except Exception as e:
        print(e)
        return

    res = {
        "address": address,
        "pubkey": pubkey,
        "mnemonic": mnemonic
    }

    return res


def delete_wallet(wallet_alias: str, passphrase: str):
    """
    clif key delete <wallet_alias>
    """
    try:
        child = pexpect.spawn("clif keys delete {}".format(wallet_alias))
        _ = child.read_nonblocking(10000, timeout=1)
        _ = child.sendline(passphrase)
        
        outs = child.read()

    except pexpect.TIMEOUT:
        raise FinishedWithError


def add_genesis_account(address: str, coin: int, stake: int):
    """
    Will deleted later

    nodef add-genesis-account <address> <initial_coin>,<initial_stake>
    """

    _ = _process_executor("nodef add-genesis-account {} {}dummy,{}stake", address, coin, stake)


def add_el_genesis_account(address: str, coin: int, stake: int):
    """
    nodef add-el-genesis-account <address> <initial_coin> <initial_stake>
    """

    _ = _process_executor("nodef add-el-genesis-account {} {} {}", address, coin, stake)

def clif_configs(chain_id: str):
    """
    clif configs for easy use
    """
    cmds = [
        "clif config chain-id {}".format(chain_id),
        "clif config output json",
        "clif config trust-node true",
        "clif config indent true"
    ]

    for cmd in cmds:
        proc = subprocess.Popen(shlex.split(cmd), stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        outs, errs = proc.communicate()
        if proc.returncode != 0:
            print(errs)
            proc.kill()
            raise FinishedWithError


def load_chainspec():
    path = os.path.join(os.environ['HOME'], ".nodef/config/manifest.toml")
    cmd = "nodef load-chainspec {}"
    _ = _process_executor(cmd, path)
    

def gentx(wallet_alias: str, passphrase: str):
    """
    nodef gentx --name <wallet_alias>
    """
    try:
        child = pexpect.spawn("nodef gentx --name {}".format(wallet_alias))
        _ = child.read_nonblocking(10000, timeout=1)
        _ = child.sendline(passphrase)
        
        outs = child.read()

    except pexpect.TIMEOUT:
        raise FinishedWithError


def collect_gentxs():
    """
    nodef collect-gentxs
    """
    _ = _process_executor("nodef collect-gentxs")

def validate_genesis():
    """
    nodef validate-genesis
    """
    _ = _process_executor("nodef validate-genesis")


def unsafe_reset_all():
    """
    nodef unsafe-reset-all
    """
    _ = _process_executor("nodef unsafe-reset-all")


def whole_cleanup():
    for item in [[".nodef", "config"], [".nodef", "data"], [".clif"]]:
        path = os.path.join(os.environ["HOME"], *item)
        shutil.rmtree(path, ignore_errors=True)


#################
## Nickname CLI
#################

def set_nickname_by_address(passphrase: str, nickname: str, address: str):
    _tx_executor("clif nickname set {} --address {}", passphrase, nickname, address)


def set_nickname_by_wallet_alias(passphrase: str, nickname: str, wallet_alias: str):
    _tx_executor("clif nickname set {} --wallet {}", passphrase, nickname, wallet_alias)


def change_to_by_address(passphrase: str, nickname: str, new_address: str, old_address: str):
    _tx_executor("clif nickname change-to {} {} --address {}", passphrase, nickname, new_address, old_address)
    

def change_to_by_wallet(passphrase: str, nickname: str, new_address: str, wallet_alias: str):
    _tx_executor("clif nickname change-to {} {} --wallet {}", passphrase, nickname, new_address, wallet_alias)


def get_address(passphrase: str, nickname: str):
    res = _process_executor("clif nickname get-address {}", nickname, need_output=True)
    return res


##################
## Hdac custom CLI
##################

def transfer_to(passphrase: str, recipient: str, amount: int, fee: int, gas_price: int, from_type: str, from_value: str):
    _tx_executor("clif hdac transfer-to {} {} {} {} --{} {}", passphrase, recipient, amount, fee, gas_price, from_type, from_value)


def bond(passphrase: str, amount: int, fee: int, gas_price: int, from_type: str, from_value: str):
    _tx_executor("clif hdac bond {} {} {} --{} {}", passphrase, amount, fee, gas_price, from_type, from_value)


def unbond(passphrase: str, amount: int, fee: int, gas_price: int, from_type: str, from_value: str):
    _tx_executor("clif hdac unbond {} {} {} --{} {}", passphrase, amount, fee, gas_price, from_type, from_value)


def get_balance(from_type: str, from_value: str):
    res = _process_executor("clif hdac getbalance --{} {}", from_type, from_value, need_output=True)
    return res


def create_validator(passphrase: str, from_value: str, pubkey: str, moniker: str, identity: str='""', website: str='""', details: str='""'):
    _tx_executor("clif hdac create-validator --from {} --pubkey {} --moniker {} --identity {} --website {} --details {}",
                      passphrase, from_value, pubkey, moniker, identity, website, details)
