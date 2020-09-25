#!/bin/bash

function deploy() {
    `nohup ./hccl-controller -v=4 >/var/log/hccl-controller 2>&1  &`
}

deploy
