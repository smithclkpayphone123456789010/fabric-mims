#!/bin/bash

set -e

echo -e "\n仅升级链码脚本（不重建网络）"

if [ $# -lt 1 ]; then
  echo "用法: ./upgrade <new_version> [chaincode_name]"
  echo "示例: ./upgrade 1.0.1 fabric-mims"
  exit 1
fi

NEW_VERSION=$1
CHAINCODE_NAME=${2:-fabric-mims}
CHANNEL_NAME=appchannel
CHAINCODE_PATH=chaincode

echo "检查操作系统类型"
if [[ `uname` == 'Darwin' ]]; then
  echo "当前操作系统是 Mac"
  export PATH=${PWD}/hyperledger-fabric-darwin-amd64-1.4.12/bin:$PATH
elif [[ `uname` == 'Linux' ]]; then
  echo "当前操作系统是 Linux"
  export PATH=${PWD}/hyperledger-fabric-linux-amd64-1.4.12/bin:$PATH
else
  echo "当前操作系统不是 Mac 或 Linux，脚本无法继续执行！"
  exit 1
fi

if ! docker ps --format '{{.Names}}' | grep -q '^cli$'; then
  echo "[Failed] 未检测到 cli 容器，请先确认网络已启动。"
  exit 1
fi

TaobaoPeer0Cli="CORE_PEER_ADDRESS=peer0.taobao.com:7051 CORE_PEER_LOCALMSPID=TaobaoMSP CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/peer/taobao.com/users/Admin@taobao.com/msp"
TaobaoPeer1Cli="CORE_PEER_ADDRESS=peer1.taobao.com:7051 CORE_PEER_LOCALMSPID=TaobaoMSP CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/peer/taobao.com/users/Admin@taobao.com/msp"
JDPeer0Cli="CORE_PEER_ADDRESS=peer0.jd.com:7051 CORE_PEER_LOCALMSPID=JDMSP CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/peer/jd.com/users/Admin@jd.com/msp"
JDPeer1Cli="CORE_PEER_ADDRESS=peer1.jd.com:7051 CORE_PEER_LOCALMSPID=JDMSP CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/peer/jd.com/users/Admin@jd.com/msp"

echo -e "\n一、在各 peer 安装链码 ${CHAINCODE_NAME}:${NEW_VERSION}（链码源码来自 cli 挂载的 ../chaincode）"
docker exec cli bash -c "$TaobaoPeer0Cli peer chaincode install -n ${CHAINCODE_NAME} -v ${NEW_VERSION} -l golang -p ${CHAINCODE_PATH}"
docker exec cli bash -c "$TaobaoPeer1Cli peer chaincode install -n ${CHAINCODE_NAME} -v ${NEW_VERSION} -l golang -p ${CHAINCODE_PATH}"
docker exec cli bash -c "$JDPeer0Cli peer chaincode install -n ${CHAINCODE_NAME} -v ${NEW_VERSION} -l golang -p ${CHAINCODE_PATH}"
docker exec cli bash -c "$JDPeer1Cli peer chaincode install -n ${CHAINCODE_NAME} -v ${NEW_VERSION} -l golang -p ${CHAINCODE_PATH}"

echo -e "\n二、在通道 ${CHANNEL_NAME} 升级链码"
docker exec cli bash -c "$TaobaoPeer0Cli peer chaincode upgrade -o orderer.qq.com:7050 -C ${CHANNEL_NAME} -n ${CHAINCODE_NAME} -l golang -v ${NEW_VERSION} -c '{\"Args\":[\"init\"]}' -P \"AND ('TaobaoMSP.member','JDMSP.member')\""

echo -e "\n三、调用 hello 验证链码可用"
docker exec cli bash -c "$TaobaoPeer0Cli peer chaincode invoke -C ${CHANNEL_NAME} -n ${CHAINCODE_NAME} -c '{\"Args\":[\"hello\"]}'"

echo -e "\n四、验证：查看已安装包与通道上已实例化版本（链码版本应为 ${NEW_VERSION}）"
docker exec cli bash -c "$TaobaoPeer0Cli peer chaincode list --installed" || true
docker exec cli bash -c "$TaobaoPeer0Cli peer chaincode list -C ${CHANNEL_NAME} --instantiated" || true

echo -e "\n[Successful] 链码升级完成。若仍报「没有该功能」，请确认虚拟机内 network/../chaincode 已含新接口，且未在升级后又执行了仅含旧源码的 install。"
