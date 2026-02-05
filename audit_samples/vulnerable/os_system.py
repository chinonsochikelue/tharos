import os
import subprocess

def run_command(cmd):
    # VULNERABLE: Command injection via os.system
    os.system("ls " + cmd)
    
    # VULNERABLE: Command injection via subprocess.run with shell=True
    subprocess.run("ping " + cmd, shell=True)

if __name__ == "__main__":
    run_command("google.com")
