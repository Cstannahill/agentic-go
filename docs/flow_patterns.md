# Additional Agentic Flow Patterns

This document sketches several orchestration flows that the current code base does not yet implement. They are ideas for future exploration and expand the capabilities outlined in the existing architecture documents.

## 1. Event Driven Pipelines

Instead of invoking a pipeline directly, a lightweight watcher listens for external events such as webhook notifications, file changes or messages on a queue. When a matching event arrives it spawns a pipeline with the event payload as the initial input. This pattern allows agents to react continuously to their environment.

**Design Considerations**

- A long running watcher goroutine subscribes to the chosen event source.
- Each received event constructs a pipeline input map and calls `ExecutePipeline` in its own goroutine.
- Backpressure can be handled with a bounded work queue so bursts of events don't overwhelm the system.
- The watcher itself is an agent so different event sources can be plugged in by configuration.

## 2. Planner–Executor Loop

A planning agent accepts a high level goal and outputs a list of pipeline steps to achieve it. The orchestrator then executes these steps dynamically. After execution a feedback agent reviews the result and either marks the goal complete or generates a new plan. This cycle continues until success criteria are met.

**Design Considerations**

- The planning agent could leverage an LLM or rules engine to convert goals into step definitions.
- Planned steps are fed back into the orchestrator to produce a temporary pipeline at runtime.
- A simple state store tracks the goal, attempted plans and final outcome for transparency.
- Safety checks should validate generated steps before execution to avoid invalid or harmful actions.

## 3. Critic Feedback Flow

For quality sensitive tasks a separate critic agent evaluates the output of another agent. The critic can request a retry with modified input, escalate to a human or approve the result. The orchestrator coordinates this handshake so the primary agent only proceeds once the critic signals success.

**Design Considerations**

- Each critical step pairs a worker agent and a critic agent in the same pipeline group.
- The critic receives the worker output and additional context such as acceptance criteria.
- If the critic marks the result unsatisfactory the step is repeated with refined parameters.
- A maximum retry count prevents infinite loops and surfaces failures to the caller.

## 4. Collaborative Multi‑Agent Workspace

Some problems benefit from multiple specialised agents working together. In this pattern agents share a lightweight data store (for example a key/value cache) and post intermediate artifacts for others to consume. The orchestrator may launch several agents at once and periodically synchronise the shared state.

**Design Considerations**

- A simple in‑memory store or database table holds shared documents or messages.
- Agents update and poll the store through dedicated helper functions.
- Synchronisation points within the pipeline ensure agents see a consistent view before continuing.
- This approach enables tasks like distributed research or synthesis where agents contribute pieces of the final answer.

## 5. Human‑in‑the‑Loop Approval

Certain actions may require explicit user confirmation. A pipeline step can pause execution and emit a request for approval via email, chat or a web UI. Once approval is granted the orchestrator resumes from the next step.

**Design Considerations**

- The approval step stores its state so the process can survive restarts.
- A separate service or manual command posts the approval which unblocks the waiting goroutine.
- Timeouts and escalation paths ensure the pipeline does not hang indefinitely.

These patterns provide inspiration for extending the orchestration layer into more autonomous and resilient workflows.
