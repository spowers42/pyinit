# pyinit

pyinit bootstraps a new Python project: it makes sure the tools the project depends on are present, collects the project's details, and creates it on disk.

## Language

**Required tool**:
An external program pyinit itself needs to create a project (uv, task); preflight checks for each one and offers to install any that are missing.
_Avoid_: dependency, prerequisite

**Location**:
Where a required tool was found, and whether that place is on the user's PATH. A tool can have a Location without being on PATH.
_Avoid_: install path

**Resolving runner**:
The means of running commands that preflight hands onward, which runs each required tool from its Location rather than relying on the user's PATH.
_Avoid_: toolchain
