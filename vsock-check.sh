#!/bin/sh
if ps -ef | grep vsock| grep -q 'kms'; 
then
  echo "matched"
else
   echo "not found"
   vsock-proxy 8000 kms.us-east-1.amazonaws.com 443 &
   echo "Restarting it again"
fi
