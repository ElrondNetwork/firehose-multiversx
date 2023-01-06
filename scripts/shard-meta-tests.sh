go build

startTest(){
  echo "starting firehosenode with screen"

  screen -L -A -m -d -S firehosenode ./integration-test.sh $1

  sleep 80

  echo "finish starting firehosenode"
  echo "starting integration tests for $1"
  if [[ "$1" == "shard" ]]
  then
    ./scripts
  else
    ./scripts --check-meta
  fi

  echo "finished integration tests for $1"

  screen -S firehosenode -X quit
}

startTest shard || exit 1
startTest metachain || exit 1

