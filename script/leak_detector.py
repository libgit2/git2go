#!/usr/bin/python3
"""Tool to assist in debugging git2go memory leaks.

In order to use, run this program as root and start the git2go binary with the
`GIT2GO_DEBUG_ALLOCATOR_LOG=/tmp/git2go_alloc` environment variable set. For
best results, make sure that the program exits and calls `git.Shutdown()` at
the end to remove most noise.
"""

import argparse
import dataclasses
import os
from typing import Dict, Iterable, TextIO, Tuple, Sequence


@dataclasses.dataclass
class Allocation:
    """A single object allocation."""

    line: str
    size: int
    ptr: int
    backtrace: Sequence[str]


def _receive_allocation_messages(
        log: TextIO) -> Iterable[Tuple[str, Allocation]]:
    for line in log:
        tokens = line.split('\t')
        message_type, ptr = tokens[:2]
        if message_type == 'D':
            yield message_type, Allocation(line='',
                                           size=0,
                                           ptr=int(ptr, 16),
                                           backtrace=())
        else:
            yield message_type, Allocation(line=tokens[3],
                                           size=int(tokens[2]),
                                           ptr=int(ptr, 16),
                                           backtrace=tuple(tokens[4:]))


@dataclasses.dataclass
class LeakSummaryEntry:
    """An entry in the leak summary."""

    allocation_count: int
    allocation_size: int
    line: str
    backtrace: Sequence[str]


def _process_leaked_allocations(
        live_allocations: Dict[int, Allocation]) -> None:
    """Print a summary of leaked allocations."""

    if not live_allocations:
        print('No leaks!')
        return

    backtraces: Dict[Sequence[str], LeakSummaryEntry] = {}
    for obj in live_allocations.values():
        if obj.backtrace not in backtraces:
            backtraces[obj.backtrace] = LeakSummaryEntry(
                0, 0, obj.line, obj.backtrace)
        backtraces[obj.backtrace].allocation_count += 1
        backtraces[obj.backtrace].allocation_size += obj.size
    print(f'{"Total size":>20} | {"Average size":>20} | '
          f'{"Allocations":>11} | Filename')
    print(f'{"":=<20}=+={"":=<20}=+={"":=<11}=+={"":=<64}')
    for entry in sorted(backtraces.values(),
                        key=lambda e: e.allocation_size,
                        reverse=True):
        print(f'{entry.allocation_size:20} | '
              f'{entry.allocation_size//entry.allocation_count:20} | '
              f'{entry.allocation_count:11} | '
              f'{entry.line}')
        for frame in entry.backtrace:
            print(f'{"":20} | {"":20} | {"":11} | {frame}')
        print(f'{"":-<20}-+-{"":-<20}-+-{"":-<11}-+-{"":-<64}')
    print()


def _handle_log(log: TextIO) -> None:
    """Parse the allocation log."""

    live_allocations: Dict[int, Allocation] = {}
    try:
        for message_type, allocation in _receive_allocation_messages(log):
            if message_type in ('A', 'R'):
                live_allocations[allocation.ptr] = allocation
            elif message_type == 'D':
                del live_allocations[allocation.ptr]
            else:
                raise Exception(f'Unknown message type "{message_type}"')
    except KeyboardInterrupt:
        pass
    _process_leaked_allocations(live_allocations)


def main() -> None:
    """Tool to assist in debugging git2go memory leaks."""

    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--pipe',
                        action='store_true',
                        help='Create a FIFO at the specified location')
    parser.add_argument('log_path',
                        metavar='PATH',
                        default='/tmp/git2go_alloc',
                        nargs='?',
                        type=str)
    args = parser.parse_args()

    if args.pipe:
        try:
            os.unlink(args.log_path)
        except FileNotFoundError:
            pass
        os.mkfifo(args.log_path)
        print('Capturing allocations, press Ctrl+C to stop...')

    with open(args.log_path, 'r') as log:
        _handle_log(log)


if __name__ == '__main__':
    main()
