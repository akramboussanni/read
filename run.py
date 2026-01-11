import subprocess
import sys
import os
from pathlib import Path

def run(command, env=None):
    print(f"→ {command}")
    result = subprocess.run(command, shell=True, env=env)
    if result.returncode != 0:
        sys.exit(result.returncode)

def main():
    # Ensure Go bin directory is in PATH
    home = Path.home()
    go_bin = home / "go" / "bin"
    env = os.environ.copy()
    
    # Add Go bin to PATH if it exists
    if go_bin.exists():
        env["PATH"] = f"{go_bin}:{env.get('PATH', '')}"
    
    if not shutil.which("swag", path=env.get("PATH")):
        print("Installing swag (requires Go to be in PATH)...")
        run("go install github.com/swaggo/swag/cmd/swag@latest", env=env)

    run("swag init -g cmd/server/main.go", env=env)
    run("go run -tags=debug cmd/server/main.go", env=env)

if __name__ == "__main__":
    import shutil
    main()
