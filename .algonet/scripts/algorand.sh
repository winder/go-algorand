ENVIRONMENT_DIR=/home/ubuntu/_algonet/environment
mkdir -p ${ENVIRONMENT_DIR}

IDENTITY_FILE=${ENVIRONMENT_DIR}/identity

echo "Network: $NETWORK"
echo "Writing Identity: $NODECFGHOST"

# Write IDENTITY
echo "export NETWORK=$NETWORK" > $IDENTITY_FILE
echo "export CHANNEL=$CHANNEL" >> $IDENTITY_FILE
echo "export NODECFGHOST=$NODECFGHOST" >> $IDENTITY_FILE
echo 'export HOSTADDRESS=$(curl --silent http://169.254.169.254/latest/meta-data/public-ipv4)' >> $IDENTITY_FILE
