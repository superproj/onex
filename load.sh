#!/bin/bash
kind load docker-image --name onex --nodes onex-worker,onex-worker2 ccr.ccs.tencentyun.com/superproj/onex-usercenter-amd64:v0.1.0
kind load docker-image --name onex --nodes onex-worker,onex-worker2 ccr.ccs.tencentyun.com/superproj/onex-apiserver-amd64:v0.1.0
kind load docker-image --name onex --nodes onex-worker,onex-worker2 ccr.ccs.tencentyun.com/superproj/onex-gateway-amd64:v0.1.0
kind load docker-image --name onex --nodes onex-worker,onex-worker2 ccr.ccs.tencentyun.com/superproj/onex-nightwatch-amd64:v0.1.0
kind load docker-image --name onex --nodes onex-worker,onex-worker2 ccr.ccs.tencentyun.com/superproj/onex-pump-amd64:v0.1.0
kind load docker-image --name onex --nodes onex-worker,onex-worker2 ccr.ccs.tencentyun.com/superproj/onex-toyblc-amd64:v0.1.0
kind load docker-image --name onex --nodes onex-worker,onex-worker2 ccr.ccs.tencentyun.com/superproj/onex-controller-manager-amd64:v0.1.0
kind load docker-image --name onex --nodes onex-worker,onex-worker2 ccr.ccs.tencentyun.com/superproj/onex-minerset-controller-amd64:v0.1.0
kind load docker-image --name onex --nodes onex-worker,onex-worker2 ccr.ccs.tencentyun.com/superproj/onex-miner-controller-amd64:v0.1.0
kind load docker-image --name onex --nodes onex-worker,onex-worker2 ccr.ccs.tencentyun.com/superproj/onex-fakeserver-amd64:v0.1.0
kind load docker-image --name onex --nodes onex-worker,onex-worker2 ccr.ccs.tencentyun.com/superproj/onex-cacheserver-amd64:v0.1.0
kind load docker-image --name onex --nodes onex-worker,onex-worker2 ccr.ccs.tencentyun.com/superproj/onex-onexctl-amd64:v0.1.0
