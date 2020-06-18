import os
import os.path
import subprocess
import shlex
import json
import re
import shutil
import time

import pexpect


from .errors import DeadDaemonException, FinishedWithError, InvalidContractRunType

def _process_executor(cmd: str, *args, need_output=False, need_json_res=True):
    res = None
    print(cmd.format(*args))
    child = pexpect.spawn(cmd.format(*args))    
    outs = child.read().decode('utf-8')

    if need_output == True:
        try:
            print(outs)
            if need_json_res == True:
                res = json.loads(outs)
        except Exception as e:
            print(e)
            raise e


    return res


def _tx_executor(cmd: str, passphrase, *args):
    try:
        print(cmd.format(*args))
        child = pexpect.spawn(cmd.format(*args))
        outs_of_child = child.read_nonblocking(30000000, timeout=3)
        outs_of_child = child.sendline('Y')
        outs_of_child = child.read_nonblocking(10000, timeout=1)
        outs_of_child = child.sendline(passphrase)
        
        outs = child.read().decode('utf-8')
        print(outs)
        try:
            tx_hash = re.search(r'"txhash": "([A-Z0-9]+)"', outs).group(1)
            success = re.search(r'"success": (true|false)', outs).group(1)
        except Exception as e:
            print(outs_of_child)
            print(e)
            raise e

    except pexpect.TIMEOUT:
        raise FinishedWithError

    return tx_hash, success == 'true'


#################
## Daemon control
#################

def run_casperlabsEE(ee_bin="../CasperLabs/execution-engine/target/release/casperlabs-engine-grpc-server",
                     socket_path=".casperlabs/.casper-node.sock") -> subprocess.Popen:
    """
    ./casperlabs-engine-grpc-server $HOME/.casperlabs/.casper-node.sock
    """
    cmd = "{} {} {}".format(ee_bin, os.path.join(os.environ['HOME'], socket_path), "-z")
    proc = subprocess.Popen(shlex.split(cmd), stdout=subprocess.DEVNULL)
    return proc


def run_node() -> subprocess.Popen:
    """
    nodef start
    """
    proc = subprocess.Popen(shlex.split("nodef start"), stdout=subprocess.DEVNULL)
    return proc


def daemon_check(proc: subprocess.Popen):
    """
    Get proc object and check whether given daemon is running or not
    """
    if proc is None:
        return True
    else:
        is_alive = proc.poll() is None
        return is_alive



#################
## Setup CLI
#################

def init_chain(moniker: str, consensus_module: str, chain_id: str) -> subprocess.Popen:
    """
    nodef init <moniker> --chain-id <chain-id>
    """
    _ = _process_executor("nodef init {} {} --chain-id {}", moniker, consensus_module, chain_id)


def copy_manifest():
    path = os.path.join(os.environ["HOME"], ".nodef/config")
    cmd = "cp ../x/executionlayer/resources/manifest.toml {}".format(path)
    _ = _process_executor(cmd, need_output=False)


def create_wallet(wallet_alias: str, passphrase: str, client_home: str = '.test_clif'):
    """
    clif key add <wallet_alias>
    """
    client_home = os.path.join(os.environ["HOME"], client_home)
    cmd = "clif keys add {} --home {}".format(wallet_alias, client_home)
    print(cmd)
    try:
        child = pexpect.spawn(cmd)
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
        try:
            # If output is not json
            address = re.search(r"address: ([a-z0-9]+)", outs).group(1)
            pubkey = re.search(r"pubkey: ([a-z0-9]+)", outs).group(1)
            mnemonic = outs.strip().split('\n')[-1]
        except Exception as e:
            print(outs)
            raise e

    except Exception as e:
        print(outs)
        raise e

    res = {
        "address": address,
        "pubkey": pubkey,
        "mnemonic": mnemonic
    }

    return res


def get_wallet_info(wallet_alias: str, client_home: str = '.test_clif'):
    """
    clif keys show <wallet_alias>
    """
    client_home = os.path.join(os.environ["HOME"], client_home)
    res = _process_executor("clif keys show {} --home {}", wallet_alias, client_home, need_output=True)
    return res


def delete_wallet(wallet_alias: str, passphrase: str, client_home: str = '.test_clif'):
    """
    clif key delete <wallet_alias>
    """
    client_home = os.path.join(os.environ["HOME"], client_home)
    try:
        child = pexpect.spawn("clif keys delete {} --home {}".format(wallet_alias, client_home))
        _ = child.read_nonblocking(10000, timeout=1)
        _ = child.sendline(passphrase)
        
        outs = child.read()

    except pexpect.TIMEOUT:
        raise FinishedWithError


def add_genesis_account(address: str, coin: str, stake: str, client_home: str = '.test_clif'):
    """
    Will deleted later

    nodef add-genesis-account <address> <initial_stake>
    """

    client_home = os.path.join(os.environ["HOME"], client_home)
    _ = _process_executor("nodef add-genesis-account {} {}stake --home-client {}", address, stake, client_home)


def add_el_genesis_account(address: str, coin: str, stake: str, client_home: str = '.test_clif'):
    """
    nodef add-el-genesis-account <address> <initial_coin> <initial_stake>
    """

    client_home = os.path.join(os.environ["HOME"], client_home)
    _ = _process_executor("nodef add-el-genesis-account {} {} {} --home-client {}", address, coin, stake, client_home)

def clif_configs(chain_id: str, client_home: str = '.test_clif'):
    """
    clif configs for easy use
    """
    client_home = os.path.join(os.environ["HOME"], client_home)
    cmds = [
        "clif config chain-id {} --home {}".format(chain_id, client_home),
        "clif config output json --home {}".format(client_home),
        "clif config trust-node true --home {}".format(client_home),
        "clif config indent true --home {}".format(client_home)
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
    

def gentx(wallet_alias: str, passphrase: str, client_home: str = '.test_clif'):
    """
    nodef gentx --name <wallet_alias>
    """
    client_home = os.path.join(os.environ["HOME"], client_home)
    try:
        child = pexpect.spawn("nodef gentx --name {} --home-client {}".format(wallet_alias, client_home))
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
    for item in [[".nodef", "config"], [".nodef", "data"], [".test_clif"]]:
        path = os.path.join(os.environ["HOME"], *item)
        shutil.rmtree(path, ignore_errors=True)


def query_tx(tx_hash, client_home: str = ".test_clif"):
    client_home = os.path.join(os.environ["HOME"], client_home)
    res = _process_executor("clif query tx {} --home {}", tx_hash, client_home, need_output=True)
    return res

def query_contract(mode, address, path, client_home: str = ".test_clif"):
    client_home = os.path.join(os.environ["HOME"], client_home)
    res = _process_executor("clif contract query {} {} {} --home {}", mode, address, path, client_home, need_output=True)
    return res

def is_tx_ok(tx_hash):
    res = query_tx(tx_hash)
    is_success = res['logs'][0]['success']
    if is_success == False or "ERROR" in res['raw_log']:
        print(res['logs'])
    return res['logs'][0]['success'] and "ERROR" not in res['raw_log']


def get_bls_pubkey_remote(remote_address):
    child = pexpect.spawn('ssh -i ~/ci_nodes.pem {} "~/go/bin/nodef tendermint show-validator"'.format(remote_address))
    outs = child.read().decode('utf-8').strip()
    return outs


#################
## Nickname CLI
#################

def set_nickname(passphrase: str, nickname: str, address: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    return _tx_executor("clif nickname set {} --from {} --node {} --home {}", passphrase, nickname, address, node, client_home)


def change_address_of_nickname(passphrase: str, nickname: str, new_address: str, old_address: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    return _tx_executor("clif nickname change-to {} {} --from {} --node {} --home {}", passphrase, nickname, new_address, old_address, node, client_home)


def get_address(nickname: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    res = _process_executor("clif nickname get-address {} --node {} --home {}", nickname, node, client_home, need_output=True)
    return res


##################
## Hdac custom CLI
##################

def transfer_to(passphrase: str, recipient: str, amount: str, fee: str, from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    return _tx_executor("clif hdac transfer-to {} {} {} --from {} --node {} --home {}", passphrase, recipient, amount, fee, from_value, node, client_home)


def bond(passphrase: str, amount: str, fee: str, from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    return _tx_executor("clif hdac bond {} {} --from {} --node {} --home {}", passphrase, amount, fee, from_value, node, client_home)


def unbond(passphrase: str, amount: str, fee: str, from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    return _tx_executor("clif hdac unbond {} {} --from {} --node {} --home {}", passphrase, amount, fee, from_value, node, client_home)

def delegate(passphrase: str, validator_address: str, amount: str, fee: str, from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    return _tx_executor("clif hdac delegate {} {} {} --from {} --node {} --home {}", passphrase, validator_address, amount, fee, from_value, node, client_home)

def undelegate(passphrase: str, validator_address: str, amount: str, fee: str, from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    return _tx_executor("clif hdac undelegate {} {} {} --from {} --node {} --home {}", passphrase, validator_address, amount, fee, from_value, node, client_home)

def redelegate(passphrase: str, src_validator_address: str, dest_validator_address: str, amount: str, fee: str, from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    return _tx_executor("clif hdac redelegate {} {} {} {} --from {} --node {} --home {}", passphrase, src_validator_address, dest_validator_address, amount, fee, from_value, node, client_home)

def vote(passphrase: str, hash: str, amount: str, fee: str, from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    return _tx_executor("clif hdac vote {} {} {} --from {} --node {} --home {}", passphrase, hash, amount, fee, from_value, node, client_home)

def unvote(passphrase: str, hash: str, amount: str, fee: str, from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    return _tx_executor("clif hdac unvote {} {} {} --from {} --node {} --home {}", passphrase, hash, amount, fee, from_value, node, client_home)

def claim_reward(passphrase: str, fee: str, from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    return _tx_executor("clif hdac claim reward {} --from {} --node {} --home {}", passphrase, fee, from_value, node, client_home)

def claim_commission(passphrase: str, fee: str, from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    return _tx_executor("clif hdac claim commission {} --from {} --node {} --home {}", passphrase, fee, from_value, node, client_home)

def get_balance(from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    res = _process_executor("clif hdac getbalance --from {} --node {} --home {}", from_value, node, client_home, need_output=True)
    return res

def get_stake(from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    res = _process_executor("clif hdac getstake --from {} --node {} --home {}", from_value, node, client_home, need_output=True)
    return res

def get_vote_user(from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    res = _process_executor("clif hdac getvote --from {} --node {} --home {}", from_value, node, client_home, need_output=True)
    return res

def get_vote_dapp(dapp_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    res = _process_executor("clif hdac getvote {} --node {} --home {}", dapp_value, node, client_home, need_output=True)
    return res

def get_validator(validator_address: str, from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    res = _process_executor("clif hdac validator {} --from {} --node {} --home {}", validator_address, from_value, node, client_home, need_output=True)
    return res

def get_delegator(validator_address: str, from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    res = _process_executor("clif hdac delegator {} --from {} --node {} --home {}", validator_address, from_value, node, client_home, need_output=True)
    return res

def get_voter(hash: str, from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    res = _process_executor("clif hdac voter {} --from {} --node {} --home {}", hash, from_value, node, client_home, need_output=True)
    return res

def get_reward(from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    res = _process_executor("clif hdac getreward --from {} --node {} --home {}", from_value, node, client_home, need_output=True)
    return res

def get_commission(from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    res = _process_executor("clif hdac getcommission --from {} --node {} --home {}", from_value, node, client_home, need_output=True)
    return res

def create_validator(passphrase: str, fee: str, from_value: str, pubkey: str, moniker: str, identity: str='""', website: str='""', details: str='""', node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    return _tx_executor("clif hdac create-validator {} --from {} --pubkey {} --moniker {} --identity {} --website {} --details {} --node {} --home {}",
                      passphrase, fee, from_value, pubkey, moniker, identity, website, details, node, client_home)

##################
## Contract exec CLI
##################

def run_contract(passphrase: str, run_type: str, run_type_value: str, args: str, fee: str, from_value: str, node: str = "tcp://localhost:26657", client_home: str = '.test_clif'):
    client_home = os.path.join(os.environ["HOME"], client_home)
    if run_type not in ["wasm", "uref", "hash", "name"]:
        raise InvalidContractRunType

    return _tx_executor("clif contract run {} {} '{}' {} --from {} --node {} --home {}",
                      passphrase, run_type, run_type_value, args, fee, from_value, node, client_home)
