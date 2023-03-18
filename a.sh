#!/bin/sh

make
sudo make install
make clean
make tests > logs.txt
anasm -d bin/funcs
micro funcs.anasm
