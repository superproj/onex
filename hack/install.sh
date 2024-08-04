#!/usr/bin/env bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#


# The root of the build/dist directory
ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${ONEX_ROOT}/hack/lib/init.sh"

function onex::install::install_cfssl()    
{    
  mkdir -p $HOME/bin/    
  wget https://github.com/cloudflare/cfssl/releases/download/v1.6.1/cfssl_1.6.1_linux_amd64 -O $HOME/bin/cfssl    
  wget https://github.com/cloudflare/cfssl/releases/download/v1.6.1/cfssljson_1.6.1_linux_amd64 -O $HOME/bin/cfssljson    
  wget https://github.com/cloudflare/cfssl/releases/download/v1.6.1/cfssl-certinfo_1.6.1_linux_amd64 -O $HOME/bin/cfssl-certinfo    
  #wget https://pkg.cfssl.org/R1.2/cfssl_linux-amd64 -O $HOME/bin/cfssl    
  #wget https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64 -O $HOME/bin/cfssljson    
  #wget https://pkg.cfssl.org/R1.2/cfssl-certinfo_linux-amd64 -O $HOME/bin/cfssl-certinfo    
  chmod +x $HOME/bin/{cfssl,cfssljson,cfssl-certinfo}    
  onex::log::info "install cfssl tools successfully"    
}   
