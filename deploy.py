import argparse
import sys
import threading
import subprocess
import time
from typing import Callable, Tuple, List, Optional, Any
import os



def uninstall_release():
    print("Uninstalling helm release")
    subprocess.run(["helm", "uninstall", "--namespace=chaos-testing", "chaos-mesh"])

def install_release():
    print("Installing release:")
    args=["helm", "install", "chaos-mesh", "helm/chaos", "--namespace=chaos-testing", "--set", "chaosDaemon.runtime=containerd", "--set", "chaosDaemon.socketPath=/run/containerd/containerd.sock"]
    ui = os.environ.get("UI", "0")
    if ui == "1":
        args.append(["--set", "dashboard.create=true"])
    subprocess.run(args)

# TODO apply/replace/ etc configurable?
def kube_apply(filename: str):
    args = ["kubectl", "apply", "-f", filename]
    subprocess.run(args)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--ui", help="Build with UI support (chaos dashboard)", action="store_true"
    )
    parser.add_argument(
        "-s",
        "--sequential",
        help="Run sequential (instead of parallel)",
        action="store_true",
    )

    return parser.parse_args()


# TODO ignore output here (optional)
def make(argument: Optional[str] = None):
    args = ["make"]
    argStr = ""
    if argument is not None:
        args.append(argument)
        argStr = argument
    else:
        ui = os.environ.get("UI", "0")
        if ui == "1":
            print("UI support enabled (chaos dashboard)")

    print("Running {}".format(" ".join(args)))
    subprocess.run(args)


def main() -> int:
    args = parse_args()

    print(f"args: {args}")
    threadHelmUninstall = threading.Thread(target=uninstall_release, args=(), kwargs={})
    threadHelmUninstall.start()

    if args.sequential:
        threadHelmUninstall.join()
        print("Helm uninstall completed")

    threadMakeGenerate = threading.Thread(target=make, args=("generate",), kwargs={})
    threadMakeGenerate.start()
    if args.sequential:
        threadMakeGenerate.join()
        print("Make generate complete")

    threadMakeYaml = threading.Thread(target=make, args=("yaml",), kwargs={})
    threadMakeYaml.start()
    if args.sequential:
        threadMakeYaml.join()
        print("Make yaml complete")

    # Here we need to make sure that `make generate` and `make yaml` have finished
    if not args.sequential:
        # If it was sequential we already waited
        print("Waiting for make generate and make yaml to finish...")
        threadMakeGenerate.join()
        threadMakeYaml.join()
        # We could be fancier and wait for the first of the two and then the other one,
        # but who cares ¯\_(ツ)_/¯
        print("Make generate and make yaml finished.. proceeding")


    # From here onwards, there is not parallelism

    if args.ui:
        os.environ["UI"] = "1"
    make()
    make("docker-push")

    install_release()
    kube_apply("manifests/")

    kube_apply("clusterrole.yaml")

if __name__ == "__main__":
    sys.exit(main())
