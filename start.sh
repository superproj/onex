 _output/platforms/darwin/arm64/onex-apiserver --etcd-servers 43.139.4.14:2379 --secure-port 52443 --client-ca-file=_output/cert/ca.pem --tls-cert-file=_output/cert/onex-apiserver.pem --tls-private-key-file=_output/cert/onex-apiserver-key.pem


_output/platforms/darwin/arm64/onex-controller-manager --kubeconfig _output/config --mysql-database=onex --mysql-host=43.139.4.14:3306 --mysql-username=onex --mysql-password='onex(#)666'
