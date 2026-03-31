#!/bin/bash

set -e

echo "[resume] 启动已部署的 Fabric 网络容器（不清理数据）"
docker-compose up -d

echo "[resume] 等待节点启动..."
sleep 5

echo "[resume] 当前容器状态："
docker-compose ps

echo "[resume] 完成。该脚本不会删除链数据、证书和配置。"
