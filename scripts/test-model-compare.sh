#!/bin/bash

#curl -XPOST -H'Content-Type: application/json' -d'{"apiVersion":"apps.onex.io/v1beta1","kind":"ModelCompare","metadata":{"name":"modelcompare0","namespace":"default"},"spec":{"displayName":"test-modelcompare","template":{"spec":{"provider":"text","sampleID":1001}},"modelIDs":[1001,1002,1003]}}' http://onex.gateway.superproj.com:32080/v1/modelcompares

#curl -XPUT -H'Content-Type: application/json' -d'{"apiVersion":"apps.onex.io/v1beta1","kind":"ModelCompare","metadata":{"name":"modelcompare0","namespace":"default"},"spec":{"displayName":"test-modelcompare-modified","template":{"spec":{"sampleID":1002}}}}' http://onex.gateway.superproj.com:32080/v1/modelcompares

#curl -XGET http://onex.gateway.superproj.com:32080/v1/modelcompares/modelcompare0

#curl -XGET http://onex.gateway.superproj.com:32080/v1/modelcompares
curl -XDELETE  http://onex.gateway.superproj.com:32080/v1/modelcompares/modelcompare0
