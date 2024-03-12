#!/usr/bin/env bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#

#Define color variables
#Feature
C_NORMAL='\033[0m';C_BOLD='\033[1m';C_DIM='\033[2m';C_UNDER='\033[4m';
C_ITALIC='\033[3m';C_NOITALIC='\033[23m';C_BLINK='\033[5m';
C_REVERSE='\033[7m';C_CONCEAL='\033[8m';C_NOBOLD='\033[22m';
C_NOUNDER='\033[24m';C_NOBLINK='\033[25m';

#Front color
C_BLACK='\033[30m';C_RED='\033[31m';C_GREEN='\033[32m';C_YELLOW='\033[33m';
C_BLUE='\033[34m';C_MAGENTA='\033[35m';C_CYAN='\033[36m';C_WHITE='\033[37m';

#background color
C_BBLACK='\033[40m';C_BRED='\033[41m';
C_BGREEN='\033[42m';C_BYELLOW='\033[43m';
C_BBLUE='\033[44m';C_BMAGENTA='\033[45m';
C_BCYAN='\033[46m';C_BWHITE='\033[47m';

# Print colors you can use
onex::color::print_color()
{
  echo
  echo -e ${bmagenta}--back-color:${normal}
  echo "bblack; bgreen; bblue; bcyan; bred; byellow; bmagenta; bwhite"
  echo
  echo -e ${red}--font-color:${normal}
  echo "black; red; green; yellow; blue; magenta; cyan; white"
  echo
  echo -e ${bold}--font:${normal}
  echo "normal; italic; reverse; nounder; bold; noitalic; conceal; noblink;
  dim; blink; nobold; under"
  echo
}

onex::color::color_print() {
  local color=$1
  shift
  # if stdout is a terminal, turn on color output.
  #   '-t' check: is a terminal?
  #   check isatty in bash https://stackoverflow.com/questions/10022323
  if [ -t 1 ]; then
    printf '\e[1;%sm%s\e[0m\n' "$color" "$*"
  else
    printf '%s\n' "$*"
  fi
}

onex::color::red()
{
  onex::color::color_print 31 "$@"
}

onex::color::green()
{
  onex::color::color_print 32 "$@"
}
