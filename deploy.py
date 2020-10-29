import argparse
import sys
import threading
import subprocess
import time
from typing import Callable, Tuple, List, Optional, Any
import os
from kubernetes import config


def print_to_output(verbose: int, out: str):
    if verbose < 2:
        return
    print(out)


def uninstall_release(verbose: int):
    print_to_output(verbose, "Uninstalling helm release")
    subprocess.run(
        ["helm", "uninstall", "--namespace=chaos-testing", "chaos-mesh"],
        stdout=(subprocess.PIPE if verbose == 3 else subprocess.DEVNULL),
        stderr=(sys.stdout.buffer if verbose > 0 else subprocess.DEVNULL)
    )


def install_release(verbose: int):
    args = [
        "helm",
        "install",
        "chaos-mesh",
        "helm/chaos-mesh",
        "--namespace=chaos-testing",
    ]
    print_to_output(verbose, "Ensuring namespace")
    ensure_namespace(verbose)
    print_to_output(verbose, "Current context check..")
    if "kind" in (config.list_kube_config_contexts()[1]["name"]):
        print_to_output(verbose, "Current context is kind")
        args = args + [
            "--set",
            "chaosDaemon.runtime=containerd",
            "--set",
            "chaosDaemon.socketPath=/run/containerd/containerd.sock",
        ]

    print_to_output(verbose, "Installing release:")
    ui = os.environ.get("UI", "0")
    if ui == "1":
        args = args + ["--set", "dashboard.create=true"]
    subprocess.run(args,
        stdout=(subprocess.PIPE if verbose == 3 else subprocess.DEVNULL),
        stderr=(sys.stdout.buffer if verbose > 0 else subprocess.DEVNULL))


# TODO apply/replace/ etc configurable?
def kube_apply(verbose: int, filename: str):
    args = ["kubectl", "apply", "-f", filename]
    subprocess.run(args,
        stdout=(subprocess.PIPE if verbose == 3 else subprocess.DEVNULL),
        stderr=(sys.stdout.buffer if verbose > 0 else subprocess.DEVNULL))


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(formatter_class=argparse.RawTextHelpFormatter)
    parser.add_argument(
        "--ui", help="Build with UI support (chaos dashboard)", action="store_true"
    )
    # TODO we might actually want to change the ImagePullPolicy here
    parser.add_argument(
        "--build-images",
        help="Build images (runs make generate, make yaml, make)",
        action="store_true",
    )
    parser.add_argument(
        "-s",
        "--sequential",
        help="Run sequential (instead of parallel)",
        action="store_true",
    )
    parser.add_argument(
        "--verboseLevel",
        "-v",
        help="""Level of verbosity (every level includes output from the previous levels):
        0 -> No Output
        1 -> Show stderr from subprocess
        2 -> Show script steps
        3 -> Show stdout from subprocess""",
        type=int,
        choices=[0, 1, 2, 3],
        metavar="",
        default=2
    )

    return parser.parse_args()


def ensure_namespace(verbose: int,ns: str = "chaos-testing"):
    args = ["kubectl", "create", "ns", ns]
    subprocess.run(args,
        stdout=(subprocess.PIPE if verbose == 3 else subprocess.DEVNULL),
        stderr=(sys.stdout.buffer if verbose > 0 else subprocess.DEVNULL))


def make(verbose: int, argument: Optional[str] = None):
    args = ["make"]
    argStr = ""
    if argument is not None:
        args.append(argument)
        argStr = argument
    else:
        ui = os.environ.get("UI", "0")
        if ui == "1":
            print_to_output(verbose, "UI support enabled (chaos dashboard)")
            print_to_output(verbose, "This might take some time...")

    print_to_output(verbose, "Running {}".format(" ".join(args)))
    subprocess.run(args,
        stdout=(subprocess.PIPE if verbose == 3 else subprocess.DEVNULL),
        stderr=(sys.stdout.buffer if verbose > 0 else subprocess.DEVNULL))


def main() -> int:
    args = parse_args()

    print_to_output(
        args.verboseLevel, f"Deploy script launched with following arguments: {args}"
    )
    threadHelmUninstall = threading.Thread(
        target=uninstall_release, args=(args.verboseLevel,), kwargs={}
    )
    threadHelmUninstall.start()

    if args.sequential:
        threadHelmUninstall.join()
        print_to_output(args.verboseLevel, "Helm uninstall completed")

    if args.build_images:
        threadMakeGenerate = threading.Thread(
            target=make, args=(args.verboseLevel, "generate",), kwargs={}
        )
        threadMakeGenerate.start()
        if args.sequential:
            threadMakeGenerate.join()
            print_to_output(args.verboseLevel, "Make generate complete")

        threadMakeYaml = threading.Thread(
            target=make, args=(args.verboseLevel, "yaml",), kwargs={}
        )
        threadMakeYaml.start()
        if args.sequential:
            threadMakeYaml.join()
            print_to_output(args.verboseLevel, "Make yaml complete")

        # Here we need to make sure that `make generate` and `make yaml` have finished
        if not args.sequential:
            # If it was sequential we already waited
            print_to_output(
                args.verboseLevel,
                "Waiting for make generate and make yaml to finish...",
            )
            threadMakeGenerate.join()
            threadMakeYaml.join()
            # We could be fancier and wait for the first of the two and then the other one,
            # but who cares ¯\_(ツ)_/¯
            print_to_output(
                args.verboseLevel, "Make generate and make yaml finished.. proceeding"
            )
    else:
        print_to_output(args.verboseLevel, "Not building images!")
    # From here onwards, there is not parallelism

    if args.ui:
        os.environ["UI"] = "1"

    if args.build_images:
        make(args.verboseLevel)
        make(args.verboseLevel, "docker-push")

    install_release(args.verboseLevel)
    kube_apply(args.verboseLevel,"manifests/")

    kube_apply(args.verboseLevel, "clusterrole.yaml")


if __name__ == "__main__":
    sys.exit(main())
