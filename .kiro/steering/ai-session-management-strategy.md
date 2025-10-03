---
inclusion: always
---

# AI Session Management Strategy

## Problem Statement

AI assistants experience context degradation and behavioral drift in long conversations, leading to:
- Loss of awareness of established guidelines
- Short-term memory issues
- Reversion to default behaviors (rushing, over-engineering)
- Ignoring systematic approaches and consultation patterns

## Hybrid Solution Strategy

### 1. Shorter, More Focused Sessions

**Session Length**: Maximum 15-20 interactions per conversation
**Session Scope**: Focus on 1-2 specific sub-tasks maximum
**Session Breaks**: Start fresh conversations for complex multi-step work

**Benefits**:
- Maintains fresh context and guideline awareness
- Prevents behavioral drift accumulation
- Reduces context window exhaustion
- Allows for clean state resets

### 2. Strategic Manual Hook Usage

**Hook Purpose**: Reset AI behavior when drifting from guidelines
**Trigger Conditions**: Use when AI starts to:
- Make changes without asking for approval
- Create multiple files/tools simultaneously
- Rush through implementation without testing
- Ignore existing working patterns
- Over-engineer solutions

**Hook Content**: Condensed reminder checklist (not full steering document)

### 3. Session State Documentation

**End-of-Session Protocol**:
- Document exactly where work left off
- Include specific next steps and current context
- Note any decisions made or approaches established
- Record any issues encountered or lessons learned

**Start-of-Session Protocol**:
- Read previous session state document
- Confirm understanding of current context
- Establish clear objectives for current session
- Review any relevant guidelines or patterns

### 4. Context-Aware Task Breakdown

**Task Sizing**: Break large tasks into session-sized chunks
**Dependency Management**: Ensure each session can complete meaningful work
**State Preservation**: Document intermediate results and decisions
**Continuity Planning**: Plan logical breakpoints for session transitions

## Implementation Guidelines

### For Users:
1. **Monitor AI behavior** for signs of drift (rushing, over-engineering, not consulting)
2. **Trigger manual hook** when behavioral issues are observed
3. **End sessions proactively** before reaching 15-20 interactions
4. **Document session state** before starting new conversations
5. **Start new sessions** with clear context and objectives

### For AI:
1. **Ask for approval** before making any code changes
2. **Work step-by-step** with testing at each stage
3. **Use systematic comparison** with existing working examples
4. **Avoid over-engineering** - build on proven solutions
5. **Consult frequently** rather than assuming requirements

## Success Metrics

- Reduced instances of over-engineering and rushed implementations
- Increased consultation and approval-seeking behavior
- Better adherence to systematic development approaches
- Fewer reverts and corrections needed
- More predictable and controlled development progress

## Trade-off Acknowledgment

This approach trades some conversation continuity for behavioral consistency and quality. The benefits of maintained guideline awareness and systematic approaches outweigh the overhead of session management and state documentation.