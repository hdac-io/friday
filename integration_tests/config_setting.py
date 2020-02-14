import os
import toml
import json
import re
from lib import cmd


def get_address_from_json(json_filename: str) -> str:
    with open(json_filename) as f:
        content = f.read()
        searched_obj = re.search(r'"memo": "([a-z0-9]+)@[0-9\.\:]+"', content)
        address = searched_obj.group(1)

        return address


def insert_address_to_seed_in_toml(toml_filename: str, home_path: str, address: str, master_ip: str):
    content = None
    with open(toml_filename, 'r') as f:
        content = toml.load(f)
        content['p2p']['seeds'] = "{}@{}:26656".format(address, master_ip)

    output_file_path = os.path.join(home_path, "config.toml")

    with open(output_file_path, "w") as f:
        toml.dump(content, f)


if __name__ == "__main__":
    HOME = os.environ.get("HOME")
    CONFIG_DIR_PATH = os.path.join(HOME, ".nodef", "config")
    GENESIS_FILE_PATH = os.path.join(CONFIG_DIR_PATH, "genesis.json")
    CONFIG_TOML_PATH = os.path.join(CONFIG_DIR_PATH, "config.toml")
    MASTER_IP = "140.238.12.186"

    print("Get node address from genesis.json")
    address = get_address_from_json(GENESIS_FILE_PATH)
    print("Address of node: {}".format(address))

    print("Insert address and IP into config.toml 'seeds' key")
    insert_address_to_seed_in_toml(CONFIG_TOML_PATH, HOME, address, MASTER_IP)
