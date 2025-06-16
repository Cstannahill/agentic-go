# Integrating Expanded Flow Patterns

This document outlines initial thoughts on how to incorporate the advanced patterns from `flow_patterns.md` into the current Go orchestrator. It does not prescribe a final design but highlights key steps for gradual adoption.

## Overview of the Existing Pipeline

Our orchestrator executes a `Pipeline` consisting of ordered `PipelineGroup`s. Each group contains multiple `PipelineStep`s that run concurrently. The orchestrator waits for all steps in a group to finish before moving to the next one. Agents are discovered via the registry in `internal/agent` and invoked using the `Execute` method. Steps may read the outputs of previous steps via simple path lookups in `StepData`.

This sequential group model is a solid foundation but does not yet cover dynamic branching, event driven triggers or feedback loops. The patterns described in `flow_patterns.md` suggest ways to extend the engine with more autonomy and resilience.

## Incremental Integration Plan

1. **Event Driven Pipelines**
   - Implement a lightweight watcher component that listens for filesystem events or webhooks. When an event occurs, it constructs the initial `StepData` map and calls `orchestrator.ExecutePipeline` in a goroutine.
   - The watcher can itself be defined as an agent so that multiple event sources may be configured using the existing registry mechanism.
   - Introduce a bounded work queue to prevent overload and provide backpressure.

2. **Planner–Executor Loop**
   - Add a new agent type `PlanningAgent` capable of turning a high level goal into a slice of `PipelineStep` definitions. These steps are fed back into the orchestrator to create a temporary pipeline at runtime.
   - Maintain a simple state store (e.g. boltDB or in‑memory map) to record goals, generated plans and outcomes. This enables transparency and repeated improvement.

3. **Critic Feedback Flow**
   - For steps where quality matters, pair the worker agent with a `CriticAgent`. After a worker step finishes, its output is passed to the critic. The critic may approve, request a retry or escalate to a human operator.
   - The orchestrator should support retry counts per step and an optional approval blocking mechanism before proceeding to the next group.

4. **Dynamic and Speculative Branching**
   - Extend `PipelineStep` with an optional `BranchLabel` field. If a step emits this label, the orchestrator chooses the matching `PipelineGroup` to run next. A default branch handles unknown labels.
   - For speculative execution, allow multiple branches to run concurrently. An aggregator agent then selects the best result and cancels the remaining branches via context cancellation.

5. **Checkpoint and Resume**
   - Periodically persist `StepData` to a durable store after each group. On startup, the orchestrator can resume from the last successful group if a previous run was interrupted.
   - This feature is particularly useful for long running or scheduled pipelines.

6. **Streaming Data Flow**
   - Augment agents to optionally return output channels instead of a single result. The orchestrator fans tokens or partial structures to downstream consumers so that steps can begin work before their predecessors fully complete.
   - Backpressure must be handled through bounded channels or rate limiting.

7. **Collaborative Multi-Agent Workspace**
   - Introduce a lightweight shared store (in-memory or simple DB) for intermediate artifacts.
   - Provide helper functions so agents can post and retrieve workspace documents.
   - Synchronisation points ensure each group has a consistent view before continuing.

8. **Human-in-the-Loop Approval**
   - Add a special step type that pauses execution pending external approval.
   - Persist waiting steps so the process can survive restarts.
   - Provide CLI or API endpoints for humans to approve or reject.

9. **Real-Time Streaming Chain**
   - Allow steps to stream partial results over channels to downstream agents.
   - Downstream agents consume tokens incrementally and optionally merge them.
   - Rate limits or bounded buffers guard against overwhelming consumers.

10. **Adaptive Model Selection**
    - Record success rate and latency metrics per model provider.
    - A policy engine selects the model at runtime using these metrics.
    - Fallback to alternate providers if the chosen model fails.

11. **Self-Healing Pipelines**
    - Add monitoring agents that watch metrics and error logs.
    - Supervisors trigger retries or step adjustments when thresholds are exceeded.
    - Escalate to operators when automatic recovery fails.

12. **Time-Scheduled Execution**
    - Implement a scheduler agent that triggers pipelines based on cron rules.
    - Maintain a history of executions and handle missed schedules.
    - Concurrency controls prevent overlapping runs.

13. **Skill Discovery Loop**
    - Provide a meta-agent that searches for missing skills or example code.
    - Retrieved artifacts are validated in a sandbox and registered if safe.
    - Track progress to avoid repeatedly learning the same capability.

14. **Reflective Self-Improvement Cycle**
    - After each run, a reflection agent analyses results and proposes changes.
    - The orchestrator can apply approved changes and rerun the pipeline.
    - Limit iterations to avoid endless tuning.

15. **Map-Reduce Fan-Out**
    - Split large datasets into chunks and run pipeline groups per chunk in parallel.
    - A reduce step aggregates partial results via a combine agent.
    - Chunking strategy and merging logic are configurable.

16. **Progressive Summarization**
    - Chain summariser agents that condense chunks of text in multiple passes.
    - Store intermediate summaries for later stages.
    - Configurable summary length and number of passes.

## Next Steps

- Prototype the event watcher and planning agent to validate the hooks needed in `Orchestrator.RunPipeline`.
- Define interfaces for critic and aggregator agents so that quality control and speculative workflows share common behaviours.
- Extend the pipeline definition schema (YAML/JSON) to express branching rules, retries and streaming indicators.

These additions will gradually evolve the existing pipeline into a more autonomous system that can adapt to real‑world scenarios while staying compatible with the current orchestration core.
