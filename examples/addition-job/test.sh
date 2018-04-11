#!/bin/bash
echo "seed batch -in addition-job-0.0.1-seed:1.0.0 -b inputs.csv -e SETTING_ONE=1 -e SETTING_TWO=2 -m MOUNT_BIN=tmpbin -m MOUNT_TMP=tmp -c cluster -cm $1"
seed cluster -in addition-job-0.0.1-seed:1.0.0 -b inputs.csv -e SETTING_ONE=1 -e SETTING_TWO=2 -m MOUNT_BIN=tmpbin -m MOUNT_TMP=tmp  -ma $1


# docker run \
# -v /data/inputs/inputs.1.txt:/data/inputs/inputs.1.txt \
# -v /data/inputs/out:/data/inputs/out \
# -v /data/tmpbin:/data/tmpbin:ro \
# -v /data/tmp:/data/tmp:rw \
# -e SETTING_ONE=1 \
# -e SETTING_TWO=2 \
# -m 16m --shm-size=128m \
# localhost:5000/addition-job-0.0.1-seed:1.0.0 \
# /data/inputs/inputs.1.txt \
# /data/out


# docker service create \
# --name add \
# --env SETTING_ONE=1 \
# --env SETTING_TWO=2 \
# --mount type=volume,src=input_vol,destination=/data/inputs \
# --mount type=volume,src=output_vol,destination=/data/out
# --mount type=volume,src=tmp_vol,destination=/data/tmp,readonly=true \
# --mount type=volume,src=tmpbin_vol,destination=/data/tmpbin \
# localhost:5000/addition-job-0.0.1-seed:1.0.0 /data/inputs/inputs.1.txt /data/out

# docker service create \
# --name add \
# --env SETTING_ONE=1 \
# --env SETTING_TWO=2 \
# --mount type=bind,src=/data/inputs/,destination=/data/inputs/ \
# --mount type=bind,src=/data/out/,destination=/data/out \
# --mount type=bind,src=/data/tmp/,destination=/data/tmp/,readonly=true \
# --mount type=bind,src=/data/tmpbin/,destination=/data/tmpbin/ \
# localhost:5000/addition-job-0.0.1-seed:1.0.0 /data/inputs/inputs.1.txt /data/out

#[
#     {
#         [
#             INPUT_FILE=/Users/emilysmith/go/.../addition-job/inputs.txt
#         ] 
#         /Users/emilysmith/go/src/github.com/ngageoint/seed-cli/examples/addition-job/batch-addition-job-0.0.1-seed_1.0.0-2018-03-15T14_23_24-04_00/1-inputs.txt
#     }
#     {
#         [
#             INPUT_FILE=/Users/emilysmith/go/src/github.com/ngageoint/seed-cli/examples/addition-job/inputs.1.txt
#         ] 
#         /Users/emilysmith/go/src/github.com/ngageoint/seed-cli/examples/addition-job/batch-addition-job-0.0.1-seed_1.0.0-2018-03-15T14_23_24-04_00/2-inputs.1.txt
#     }
#     {
#         [
#             INPUT_FILE=/Users/emilysmith/go/src/github.com/ngageoint/seed-cli/examples/addition-job/inputs.2.txt
#         ]
#          /Users/emilysmith/go/src/github.com/ngageoint/seed-cli/examples/addition-job/batch-addition-job-0.0.1-seed_1.0.0-2018-03-15T14_23_24-04_00/3-inputs.2.txt
#     }
#     {
#         [
#             INPUT_FILE=/Users/emilysmith/go/src/github.com/ngageoint/seed-cli/examples/addition-job/inputs.3.txt
#         ] 
#         /Users/emilysmith/go/src/github.com/ngageoint/seed-cli/examples/addition-job/batch-addition-job-0.0.1-seed_1.0.0-2018-03-15T14_23_24-04_00/4-inputs.3.txt
#     }
# ]
