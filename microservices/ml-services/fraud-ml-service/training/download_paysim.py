from __future__ import annotations

import argparse
import shutil
from pathlib import Path

import kagglehub


def find_csv(dataset_dir: Path) -> Path:
    candidates = sorted(dataset_dir.rglob("*.csv"))
    if not candidates:
        raise FileNotFoundError(f"No CSV files found under {dataset_dir}")
    for candidate in candidates:
        if "paysim" in candidate.name.lower():
            return candidate
    return candidates[0]


def main() -> None:
    parser = argparse.ArgumentParser(description="Download the public PaySim dataset through kagglehub.")
    parser.add_argument(
        "--output",
        type=Path,
        default=Path(__file__).resolve().parent / "paysim.csv",
        help="Target CSV path inside the repo.",
    )
    args = parser.parse_args()

    dataset_path = Path(kagglehub.dataset_download("ealaxi/paysim1"))
    source_csv = find_csv(dataset_path)

    args.output.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(source_csv, args.output)

    print(f"Downloaded dataset directory: {dataset_path}")
    print(f"Selected CSV: {source_csv}")
    print(f"Copied CSV to: {args.output}")


if __name__ == "__main__":
    main()
