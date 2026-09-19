#!/bin/bash
# Test script for lem-in.
set -u
cd "$(dirname "$0")"

pass=0; fail=0
check_turns() { # file max_turns
  out=$(go run . "tests/$1" 2>&1); code=$?
  if [ $code -ne 0 ]; then echo "FAIL $1: exit $code"; echo "$out" | head -3; fail=$((fail+1)); return; fi
  n=$(echo "$out" | grep -c "^L")
  if [ "$n" -le "$2" ]; then echo "PASS $1: $n turns (<= $2)"; pass=$((pass+1));
  else echo "FAIL $1: $n turns (> $2)"; fail=$((fail+1)); fi
}
check_error() { # file
  out=$(go run . "tests/$1" 2>&1); code=$?
  if [ $code -eq 0 ]; then echo "FAIL $1: expected error"; fail=$((fail+1)); return; fi
  if echo "$out" | grep -q "ERROR: invalid data format"; then echo "PASS $1: got ERROR"; pass=$((pass+1));
  else echo "FAIL $1: wrong message: $out"; fail=$((fail+1)); fi
}

echo "== vet =="
go vet ./... || exit 1

echo "== examples (turn limits from audit) =="
check_turns example00.txt 6
check_turns example01.txt 8
check_turns example02.txt 11
check_turns example03.txt 6
check_turns example04.txt 6
check_turns example05.txt 8

echo "== bad examples =="
check_error badexample00.txt
check_error badexample01.txt

echo "== perf =="
time go run . tests/example06.txt > /dev/null
sed '1s/.*/100/' tests/example06.txt > /tmp/e06_100.txt
echo -n "example06 with 100 ants: "; time go run . /tmp/e06_100.txt > /dev/null
sed '1s/.*/1000/' tests/example07.txt > /tmp/e07_1000.txt
echo -n "example07 with 1000 ants: "; time go run . /tmp/e07_1000.txt > /dev/null

echo "== $pass passed, $fail failed =="
[ "$fail" -eq 0 ]
