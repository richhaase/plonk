#!/usr/bin/env python
"""
Simple UI/UX comparison tool for Plonk.

Compares output between installed plonk and compiled plonk, showing diffs
for human review. The user decides if changes are expected or not.
"""

import argparse
import json
import os
import subprocess
import sys
import tempfile
import shutil
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Tuple
import difflib
import yaml

# Colors for terminal output
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    CYAN = '\033[0;36m'
    NC = '\033[0m'  # No Color


class PlonkCompare:
    def __init__(self, verbose: bool = False):
        self.verbose = verbose
        self.script_dir = Path(__file__).parent
        self.project_root = self.script_dir.parent.parent
        self.results_dir = self.script_dir / 'results'
        self.results_dir.mkdir(exist_ok=True)

    def run(self, update_baseline: bool = False, filter_pattern: Optional[str] = None,
            skip_build: bool = False):
        """Run the comparison."""

        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        session_dir = self.results_dir / timestamp
        session_dir.mkdir(exist_ok=True)

        print(f"{Colors.BLUE}=== Plonk UI/UX Comparison Tool ==={Colors.NC}")
        print(f"Session: {timestamp}")
        print()

        # Check prerequisites
        if not self._check_prerequisites(skip_build):
            return False

        # Build if needed
        if not skip_build and not update_baseline:
            self._build_plonk()

        # Load test scenarios
        scenarios = self._load_scenarios(filter_pattern)
        print(f"Running {len(scenarios)} test scenarios...")
        print()

        # Save current environment
        original_plonk_dir = os.environ.get('PLONK_DIR', '')

        try:
            # Create safe test environment
            with tempfile.TemporaryDirectory() as temp_dir:
                test_plonk_dir = Path(temp_dir) / 'plonk-test'

                # Copy user's plonk config if it exists
                if original_plonk_dir and Path(original_plonk_dir).exists():
                    shutil.copytree(original_plonk_dir, test_plonk_dir)
                else:
                    default_plonk_dir = Path.home() / '.config' / 'plonk'
                    if default_plonk_dir.exists():
                        shutil.copytree(default_plonk_dir, test_plonk_dir)
                    else:
                        test_plonk_dir.mkdir(parents=True)

                # Set test environment
                os.environ['PLONK_DIR'] = str(test_plonk_dir)

                if update_baseline:
                    # Just capture baseline and save it
                    print(f"{Colors.BLUE}Capturing baseline from installed plonk...{Colors.NC}")
                    baseline = self._capture_outputs('plonk', scenarios, session_dir / 'baseline')

                    # Save as current baseline
                    baseline_file = self.script_dir / 'baseline.json'
                    with open(baseline_file, 'w') as f:
                        json.dump(baseline, f, indent=2)

                    print(f"{Colors.GREEN}✓ Baseline updated{Colors.NC}")
                    print(f"Baseline saved to: {baseline_file}")

                else:
                    # Load baseline
                    baseline_file = self.script_dir / 'baseline.json'
                    if not baseline_file.exists():
                        print(f"{Colors.RED}No baseline found. Run with --update-baseline first.{Colors.NC}")
                        return False

                    with open(baseline_file) as f:
                        baseline = json.load(f)

                    # Capture current outputs
                    print(f"{Colors.BLUE}Testing compiled plonk...{Colors.NC}")
                    current = self._capture_outputs(
                        str(self.project_root / 'bin' / 'plonk'),
                        scenarios,
                        session_dir / 'current'
                    )

                    # Compare and show results
                    self._compare_and_report(baseline, current, session_dir)

        finally:
            # Restore environment
            if original_plonk_dir:
                os.environ['PLONK_DIR'] = original_plonk_dir
            else:
                os.environ.pop('PLONK_DIR', None)

        return True

    def _check_prerequisites(self, skip_build: bool) -> bool:
        """Check that required tools are available."""
        missing = []

        if not shutil.which('plonk'):
            missing.append('plonk (installed version)')

        if not skip_build:
            if not shutil.which('just'):
                missing.append('just (build tool)')

        if missing:
            print(f"{Colors.RED}Missing prerequisites:{Colors.NC}")
            for item in missing:
                print(f"  - {item}")
            return False

        return True

    def _build_plonk(self):
        """Build plonk."""
        print(f"{Colors.BLUE}Building plonk...{Colors.NC}")
        result = subprocess.run(
            ['just', 'build'],
            cwd=self.project_root,
            capture_output=True,
            text=True
        )
        if result.returncode != 0:
            print(f"{Colors.RED}Build failed:{Colors.NC}")
            print(result.stderr)
            sys.exit(1)
        print(f"{Colors.GREEN}✓ Build complete{Colors.NC}")
        print()

    def _load_scenarios(self, filter_pattern: Optional[str]) -> List[Dict]:
        """Load test scenarios from configuration."""
        scenarios_file = self.script_dir / 'scenarios.yaml'

        # Create default scenarios if file doesn't exist
        if not scenarios_file.exists():
            self._create_default_scenarios(scenarios_file)

        with open(scenarios_file) as f:
            all_scenarios = yaml.safe_load(f)

        # Filter if requested
        if filter_pattern:
            scenarios = [s for s in all_scenarios
                        if filter_pattern in s.get('name', '') or
                           filter_pattern in s.get('command', '')]
        else:
            scenarios = all_scenarios

        return scenarios

    def _create_default_scenarios(self, scenarios_file: Path):
        """Create default test scenarios."""
        default_scenarios = [
            # Basic commands
            {'name': 'help', 'command': '--help'},
            {'name': 'version', 'command': '--version'},

            # List commands
            {'name': 'list-all', 'command': 'list'},
            {'name': 'list-packages', 'command': 'list --packages'},
            {'name': 'list-dotfiles', 'command': 'list --dotfiles'},

            # Status
            {'name': 'status', 'command': 'status'},
            {'name': 'status-verbose', 'command': 'status -v'},

            # Info commands
            {'name': 'info-brew', 'command': 'info brew'},
            {'name': 'info-npm', 'command': 'info npm'},

            # Manager commands
            {'name': 'managers', 'command': 'managers'},

            # Config command
            {'name': 'config', 'command': 'config'},

            # Error cases (these should fail gracefully)
            {'name': 'error-unknown-command', 'command': 'nonexistent'},
            {'name': 'error-add-no-args', 'command': 'add'},
            {'name': 'error-remove-no-args', 'command': 'remove'},

            # Dry run commands (safe to test)
            {'name': 'add-dry-run', 'command': 'add --dry-run vim'},
            {'name': 'remove-dry-run', 'command': 'remove --dry-run vim'},
            {'name': 'apply-dry-run', 'command': 'apply --dry-run'},
        ]

        scenarios_file.parent.mkdir(exist_ok=True)
        with open(scenarios_file, 'w') as f:
            yaml.dump(default_scenarios, f, default_flow_style=False)

        print(f"{Colors.YELLOW}Created default scenarios file: {scenarios_file}{Colors.NC}")
        print("You can edit this file to add more test cases.")
        print()

    def _capture_outputs(self, plonk_path: str, scenarios: List[Dict],
                        output_dir: Path) -> Dict[str, Dict]:
        """Capture outputs for all scenarios."""
        output_dir.mkdir(exist_ok=True)
        results = {}

        for scenario in scenarios:
            name = scenario['name']
            command = scenario['command']

            if self.verbose:
                print(f"  Running: plonk {command}")

            # Build full command
            cmd_parts = [plonk_path] + command.split()

            # Run command
            start_time = datetime.now()
            result = subprocess.run(
                cmd_parts,
                capture_output=True,
                text=True,
                env=os.environ.copy()
            )
            duration = (datetime.now() - start_time).total_seconds()

            # Save outputs
            output_data = {
                'command': command,
                'stdout': result.stdout,
                'stderr': result.stderr,
                'exit_code': result.returncode,
                'duration': duration
            }

            results[name] = output_data

            # Save to files for debugging
            scenario_dir = output_dir / name
            scenario_dir.mkdir(exist_ok=True)

            (scenario_dir / 'stdout.txt').write_text(result.stdout)
            (scenario_dir / 'stderr.txt').write_text(result.stderr)
            (scenario_dir / 'info.json').write_text(json.dumps({
                'command': command,
                'exit_code': result.returncode,
                'duration': duration
            }, indent=2))

        return results

    def _compare_and_report(self, baseline: Dict, current: Dict, session_dir: Path):
        """Compare outputs and generate report."""
        print()
        print(f"{Colors.BLUE}=== Comparison Results ==={Colors.NC}")
        print()

        differences = []
        identical = []

        for name, baseline_data in baseline.items():
            if name not in current:
                print(f"{Colors.RED}✗ {name}: Missing in current run{Colors.NC}")
                differences.append(name)
                continue

            current_data = current[name]
            has_diff = False

            # Compare stdout
            if baseline_data['stdout'] != current_data['stdout']:
                has_diff = True

            # Compare stderr
            if baseline_data['stderr'] != current_data['stderr']:
                has_diff = True

            # Compare exit code
            if baseline_data['exit_code'] != current_data['exit_code']:
                has_diff = True

            if has_diff:
                print(f"{Colors.YELLOW}≠ {name}: Output differs{Colors.NC}")
                differences.append(name)

                # Save diff
                diff_file = session_dir / f"{name}.diff"
                self._save_diff(name, baseline_data, current_data, diff_file)

                if self.verbose:
                    print(f"  Diff saved to: {diff_file}")
            else:
                print(f"{Colors.GREEN}✓ {name}: Identical{Colors.NC}")
                identical.append(name)

        # Summary
        print()
        print(f"{Colors.BLUE}Summary:{Colors.NC}")
        print(f"  Total scenarios: {len(baseline)}")
        print(f"  Identical: {Colors.GREEN}{len(identical)}{Colors.NC}")
        print(f"  Different: {Colors.YELLOW}{len(differences)}{Colors.NC}")

        if differences:
            print()
            print(f"{Colors.YELLOW}Review the diffs to determine if changes are expected:{Colors.NC}")
            print(f"  - Refactoring: Changes indicate potential issues")
            print(f"  - Enhancement: Changes should match your new features")
            print()
            print(f"Diffs saved in: {session_dir}")

            # Create summary report
            summary_file = session_dir / 'summary.md'
            self._create_summary_report(baseline, current, differences, summary_file)
            print(f"Full report: {summary_file}")

    def _save_diff(self, name: str, baseline: Dict, current: Dict, diff_file: Path):
        """Save a unified diff for a scenario."""
        with open(diff_file, 'w') as f:
            f.write(f"Scenario: {name}\n")
            f.write(f"Command: plonk {baseline['command']}\n")
            f.write("=" * 80 + "\n\n")

            # Exit code diff
            if baseline['exit_code'] != current['exit_code']:
                f.write(f"Exit Code:\n")
                f.write(f"  Baseline: {baseline['exit_code']}\n")
                f.write(f"  Current:  {current['exit_code']}\n\n")

            # Stdout diff
            if baseline['stdout'] != current['stdout']:
                f.write("STDOUT Diff:\n")
                f.write("-" * 40 + "\n")
                diff = difflib.unified_diff(
                    baseline['stdout'].splitlines(keepends=True),
                    current['stdout'].splitlines(keepends=True),
                    fromfile='baseline',
                    tofile='current'
                )
                f.writelines(diff)
                f.write("\n")

            # Stderr diff
            if baseline['stderr'] != current['stderr']:
                f.write("STDERR Diff:\n")
                f.write("-" * 40 + "\n")
                diff = difflib.unified_diff(
                    baseline['stderr'].splitlines(keepends=True),
                    current['stderr'].splitlines(keepends=True),
                    fromfile='baseline',
                    tofile='current'
                )
                f.writelines(diff)

    def _create_summary_report(self, baseline: Dict, current: Dict,
                              differences: List[str], summary_file: Path):
        """Create a markdown summary report."""
        with open(summary_file, 'w') as f:
            f.write("# Plonk UI/UX Comparison Report\n\n")
            f.write(f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n")

            f.write("## Summary\n\n")
            f.write(f"- Total scenarios: {len(baseline)}\n")
            f.write(f"- Identical: {len(baseline) - len(differences)}\n")
            f.write(f"- Different: {len(differences)}\n\n")

            if differences:
                f.write("## Scenarios with Differences\n\n")
                for name in differences:
                    f.write(f"### {name}\n\n")
                    f.write(f"Command: `plonk {baseline[name]['command']}`\n\n")

                    if name in current:
                        if baseline[name]['exit_code'] != current[name]['exit_code']:
                            f.write(f"**Exit code changed:** ")
                            f.write(f"{baseline[name]['exit_code']} → {current[name]['exit_code']}\n\n")

                        if baseline[name]['stdout'] != current[name]['stdout']:
                            f.write("**Stdout changed** (see diff file)\n\n")

                        if baseline[name]['stderr'] != current[name]['stderr']:
                            f.write("**Stderr changed** (see diff file)\n\n")
                    else:
                        f.write("**Scenario missing in current run**\n\n")


def main():
    parser = argparse.ArgumentParser(
        description='Compare UI/UX between installed and compiled plonk',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # First time setup - capture baseline from installed plonk
  %(prog)s --update-baseline

  # Test after making changes
  %(prog)s

  # Test specific scenarios
  %(prog)s -f list

  # Skip building (if already built)
  %(prog)s --skip-build

  # Verbose output
  %(prog)s -v
"""
    )

    parser.add_argument('--update-baseline', action='store_true',
                        help='Update baseline from installed plonk')
    parser.add_argument('-f', '--filter', help='Filter scenarios by pattern')
    parser.add_argument('--skip-build', action='store_true',
                        help='Skip building plonk')
    parser.add_argument('-v', '--verbose', action='store_true',
                        help='Show detailed output')

    args = parser.parse_args()

    compare = PlonkCompare(verbose=args.verbose)
    success = compare.run(
        update_baseline=args.update_baseline,
        filter_pattern=args.filter,
        skip_build=args.skip_build
    )

    if not success:
        sys.exit(1)


if __name__ == '__main__':
    main()
