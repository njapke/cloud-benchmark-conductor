#!/usr/bin/env bash

set -euo pipefail
shopt -s globstar

newSuffix="__100000__0_00000__0_00000.dot"
# suffix of dotfiles: __100000__0_0__0_0.dot
for filename in $1/**/*.dot; do
  filename_without_suffix="${filename%.dot}"
    echo "checking: $filename"
    if [[ $filename != *"$newSuffix" ]]; then
      newFilename="${filename_without_suffix}${newSuffix}"
      echo "renaming $filename to $newFilename"
      mv "$filename" "$newFilename"
    fi
done
