#!/usr/bin/env bash
# Simulates a pyinit session for demo recording.
# Run via: asciinema rec --command "bash demo/simulate.sh" demo/demo.cast

BOLD='\033[1m'
DIM='\033[2m'
CYAN='\033[38;5;39m'
GREEN='\033[38;5;76m'
PURPLE='\033[38;5;141m'
YELLOW='\033[38;5;220m'
RESET='\033[0m'

type_out() {
  local text="$1"
  local delay="${2:-0.06}"
  for (( i=0; i<${#text}; i++ )); do
    printf "%s" "${text:$i:1}"
    sleep "$delay"
  done
}

clear
sleep 0.4

# ── Banner ──────────────────────────────────────────────────────────────────
printf "${BOLD}pyinit${RESET} — bootstrap a new Python project\n\n"
sleep 0.6

# ── Group 1: Project name + Description ─────────────────────────────────────
printf "${PURPLE}  Project name${RESET}\n"
printf "  ${DIM}lowercase, hyphens and underscores allowed${RESET}\n"
printf "  ${CYAN}>${RESET} "
sleep 0.4
type_out "my-awesome-lib"
printf "\n\n"
sleep 0.3

printf "${PURPLE}  Description${RESET}\n"
printf "  ${CYAN}>${RESET} "
sleep 0.3
type_out "A Python library for doing awesome things"
printf "\n\n"
sleep 0.5

# Clear and redraw confirmed group
printf "\033[6A\033[0J"
printf "${GREEN}  ✓ Project name${RESET}      ${DIM}my-awesome-lib${RESET}\n"
printf "${GREEN}  ✓ Description${RESET}       ${DIM}A Python library for doing awesome things${RESET}\n\n"
sleep 0.4

# ── Group 2: Author ──────────────────────────────────────────────────────────
printf "${PURPLE}  Author name${RESET}\n"
printf "  ${CYAN}>${RESET} "
sleep 0.3
type_out "Scott Powers"
printf "\n\n"
sleep 0.3

printf "${PURPLE}  Author email${RESET}\n"
printf "  ${CYAN}>${RESET} "
sleep 0.3
type_out "scott@example.com"
printf "\n\n"
sleep 0.5

printf "\033[6A\033[0J"
printf "${GREEN}  ✓ Author name${RESET}       ${DIM}Scott Powers${RESET}\n"
printf "${GREEN}  ✓ Author email${RESET}      ${DIM}scott@example.com${RESET}\n\n"
sleep 0.4

# ── Group 3: Python version + Output dir ─────────────────────────────────────
printf "${PURPLE}  Python version${RESET}\n"
printf "  ${CYAN}>${RESET} ${DIM}3.12${RESET}"
sleep 0.8
printf "\n\n"

printf "${PURPLE}  Output directory${RESET}\n"
printf "  ${DIM}Parent directory where the project folder will be created${RESET}\n"
printf "  ${CYAN}>${RESET} "
sleep 0.3
type_out "~/Development"
printf "\n\n"
sleep 0.5

printf "\033[8A\033[0J"
printf "${GREEN}  ✓ Python version${RESET}    ${DIM}3.12${RESET}\n"
printf "${GREEN}  ✓ Output directory${RESET}  ${DIM}~/Development${RESET}\n\n"
sleep 0.4

# ── Scaffold output ───────────────────────────────────────────────────────────
printf "Creating project ${BOLD}\"my-awesome-lib\"${RESET} in ~/Development...\n\n"
sleep 0.4

printf "  Running ${CYAN}uv init${RESET}...\n"
sleep 0.6
printf "  ${DIM}Initialized project \`my-awesome-lib\` at \`~/Development/my-awesome-lib\`${RESET}\n"
sleep 0.4

printf "  Running ${CYAN}uv venv${RESET}...\n"
sleep 0.8
printf "  ${DIM}Using CPython 3.12.10${RESET}\n"
printf "  ${DIM}Creating virtual environment at: .venv${RESET}\n"
sleep 0.5

# ── Success ───────────────────────────────────────────────────────────────────
printf "\n${GREEN}${BOLD}Done!${RESET} Your project is ready at ~/Development/my-awesome-lib\n\n"
sleep 0.3

printf "Next steps:\n"
printf "  ${CYAN}cd my-awesome-lib${RESET}\n"
sleep 0.15
printf "  ${CYAN}task install${RESET}    ${DIM}# install dependencies${RESET}\n"
sleep 0.15
printf "  ${CYAN}task check${RESET}      ${DIM}# lint + test${RESET}\n"
sleep 1.5
