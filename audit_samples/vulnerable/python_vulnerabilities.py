import os
import pickle
import yaml
import subprocess
import requests

def vulnerable_functions():
    # 1. Insecure Deserialization
    data = b"cos\nsystem\n(S'ls'\ntR."
    pickle.loads(data)
    
    # 2. Unsafe YAML Load
    yaml.load("!!python/object/apply:os.system ['ls']", Loader=yaml.Loader)

    # 3. Command Injection
    cmd = input("Enter command: ")
    os.system("echo " + cmd)
    subprocess.call("echo " + cmd, shell=True)

    # 4. Unsafe Eval/Exec
    user_input = "1 + 1"
    eval(user_input)
    exec("print('hello')")

    # 5. Insecure Requests
    requests.get("https://example.com", verify=False)

    # 6. Hardcoded Secret
    AWS_SECRET = "AKIA_FAKE_SECRET_KEY_FOR_TESTING_1234567890"
    print(AWS_SECRET)

if __name__ == "__main__":
    vulnerable_functions()
