CURRENT_DIR=$(pwd)
WORKING_DIR="$CURRENT_DIR"
TESTNET_DIR=$WORKING_DIR/testnet
TESTNET_OUTPUT_DIR=$TESTNET_DIR/testnet-local
SCRIPTS_DIR=elrond-go/scripts/testnet
VARIABLES_PATH=$SCRIPTS_DIR/variables.sh
OBSERVERS_PATH=$SCRIPTS_DIR/include/observers.sh
VALIDATORS_PATH=$SCRIPTS_DIR/include/validators.sh
ENABLE_EPOCH_DIR=$TESTNET_DIR/elrond-go/cmd/node/config/enableEpochs.toml
SYSTEM_SC_CONFIG_DIR=$TESTNET_DIR/elrond-go/cmd/node/config/systemSmartContractsConfig.toml
SANDBOX_NAME=sandbox

cloneDependencies(){
  if [ -d "$TESTNET_DIR" ]; then
    rm -rf $TESTNET_DIR
  fi

  mkdir "$TESTNET_DIR"

  git clone https://github.com/ElrondNetwork/elrond-go "$TESTNET_DIR/elrond-go"
  cd $TESTNET_DIR/elrond-go
  git checkout 3fb9427ba444b5d5dc6072713c45756b4e45d79c
  cd ../..

  git clone https://github.com/ElrondNetwork/elrond-deploy-go "$TESTNET_DIR/elrond-deploy-go"
  git clone https://github.com/ElrondNetwork/elrond-proxy-go "$TESTNET_DIR/elrond-proxy-go"
}

testnetRemove(){
  rm -rf "$TESTNET_OUTPUT_DIR"
}

testnetSetup(){
  sed -i 's/TransactionSignedWithTxHashEnableEpoch =.*/TransactionSignedWithTxHashEnableEpoch = 0/' "$ENABLE_EPOCH_DIR"
  sed -i 's/BuiltInFunctionsEnableEpoch =.*/BuiltInFunctionsEnableEpoch = 0/' "$ENABLE_EPOCH_DIR"
  sed -i 's/ESDTEnableEpoch =.*/ESDTEnableEpoch = 0/' "$ENABLE_EPOCH_DIR"
  sed -i 's/ESDTMultiTransferEnableEpoch =.*/ESDTMultiTransferEnableEpoch = 0/' "$ENABLE_EPOCH_DIR"
  sed -i 's/ESDTTransferRoleEnableEpoch =.*/ESDTTransferRoleEnableEpoch = 0/' "$ENABLE_EPOCH_DIR"
  sed -i 's/MetaESDTSetEnableEpoch =.*/MetaESDTSetEnableEpoch = 0/' "$ENABLE_EPOCH_DIR"
  sed -i 's/ESDTRegisterAndSetAllRolesEnableEpoch =.*/ESDTRegisterAndSetAllRolesEnableEpoch = 0/' "$ENABLE_EPOCH_DIR"
  sed -i 's/ESDTRegisterAndSetAllRolesEnableEpoch =.*/ESDTRegisterAndSetAllRolesEnableEpoch = 0/' "$ENABLE_EPOCH_DIR"
  sed -i 's/BaseIssuingCost =.*/BaseIssuingCost = "50000000000000000"/' "$SYSTEM_SC_CONFIG_DIR"

  mkdir "$TESTNET_OUTPUT_DIR"
  cd "$TESTNET_OUTPUT_DIR" && \
    (ln -s "$TESTNET_DIR"/elrond-go elrond-go && \
    ln -s "$TESTNET_DIR"/elrond-deploy-go elrond-deploy-go && \
    ln -s "$TESTNET_DIR"/elrond-proxy-go elrond-proxy-go)
}

testnetPrereq(){
  cd "$TESTNET_OUTPUT_DIR" && \
    bash $SCRIPTS_DIR/prerequisites.sh && \
    echo -e "export TESTNETDIR=$TESTNET_OUTPUT_DIR/$SANDBOX_NAME" > $SCRIPTS_DIR/local.sh && \
    bash $SCRIPTS_DIR/config.sh
}

testnetUpdateVariables(){
  sed -i 's/export SHARDCOUNT=.*/export SHARDCOUNT=1/' $VARIABLES_PATH
  sed -i 's/SHARD_VALIDATORCOUNT=.*/SHARD_VALIDATORCOUNT=1/' $VARIABLES_PATH
  sed -i 's/SHARD_OBSERVERCOUNT=.*/SHARD_OBSERVERCOUNT=1/' $VARIABLES_PATH
  sed -i 's/SHARD_CONSENSUS_SIZE=.*/SHARD_CONSENSUS_SIZE=$SHARD_VALIDATORCOUNT/' $VARIABLES_PATH
  sed -i 's/META_VALIDATORCOUNT=.*/META_VALIDATORCOUNT=1/' $VARIABLES_PATH
  sed -i 's/META_OBSERVERCOUNT=.*/META_OBSERVERCOUNT=1/' $VARIABLES_PATH
  sed -i 's/META_CONSENSUS_SIZE=.*/META_CONSENSUS_SIZE=$META_VALIDATORCOUNT/' $VARIABLES_PATH
  sed -i 's/export NODE_DELAY=.*/export NODE_DELAY=30/' $VARIABLES_PATH

  sed -i 's/EXTRA_OBSERVERS_FLAGS.*/EXTRA_OBSERVERS_FLAGS --firehose-enabled --operation-mode db-lookup-extension"/' $OBSERVERS_PATH
  sed -i 's/config_validator.toml/config_validator.toml --firehose-enabled --operation-mode db-lookup-extension/' $VALIDATORS_PATH
}

testnetNew(){
  cloneDependencies
  testnetRemove
  testnetSetup
  testnetUpdateVariables
  testnetPrereq
}

testnetStart(){
  cd "$TESTNET_DIR" && \
    ./elrond-go/scripts/testnet/start.sh
}

testnetReset(){
  cd "$TESTNET_DIR" && \
    ./elrond-go/scripts/testnet/reset.sh
}

testnetStop(){
  cd "$TESTNET_DIR" && \
    ./elrond-go/scripts/testnet/stop.sh
}

echoOptions(){
  echo "ERROR!!! Please choose one of the following parameters:
  - new to create a new testnet
  - start to start the testnet
  - reset to reset the testnet
  - stop to stop the testnet"
}

main(){
  if [ $# -eq 1 ]; then
    case "$1" in
      new)
        testnetNew ;;
      start)
        testnetStart ;;
      reset)
        testnetReset ;;
      stop)
        testnetStop ;;
      *)
        echoOptions ;;
      esac
  else
    echoOptions
  fi
}

main "$@"
