#!/usr/bin/env bash
echo $(pwd)
set -e
set -u
set -x
# make kafka proto file
repo_root=.
out_dir=.
proto_file=${repo_root}/rpc/western_queen.proto

protoc  ${proto_file} --go_out=plugins=grpc:${out_dir}