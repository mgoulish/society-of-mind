# vision-society

An experimental non-neural emergent AI system for machine vision, loosely inspired by Marvin Minsky’s *Society of Mind* and classic blackboard architectures.

## Current Status

**v0.1** — Minimal working skeleton.

The system currently demonstrates the core interaction loop:

- A `SyntheticBlobEmitter` posts synthetic `RedBlob` Abstractions (with numeric size/position data).
- A `SimpleBlobTracker` subscribes to `RedBlob`, maintains simple identity and history, and posts `TrackedBlob` Abstractions.
- All activity is logged in structured JSON.
- The system runs cleanly with graceful shutdown.

This version proves that the basic blackboard + autonomous Abstractor model works in Go.

## Goals

The long-term goal is to explore **emergent adaptive behavior** that arises from the interactions between many simple, loosely coupled components rather than from neural networks or hand-written pipelines.

We are particularly interested in behaviors such as:

- Focus of attention and suppression
- Object identity and tracking across frames
- Incremental hypothesis building
- On-demand specialization (dynamic creation of Abstractors)
- Robustness through redundancy and multiple paths

## Quick Start

```bash
git clone https://github.com/mgoulish/vision-society.git
cd vision-society
go run ./cmd/vision-society


