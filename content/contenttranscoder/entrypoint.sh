#!/bin/bash

# check that all needed env vars are set
: "${REDIS_USERNAME:?"REDIS_USERNAME is unset"}"
: "${REDIS_PASSWORD:?"REDIS_PASSWORD is unset"}"
: "${REDIS_HOST:?"REDIS_HOST is unset"}"
: "${REDIS_PORT:?"REDIS_PORT is unset"}"
: "${REDIS_JOB_QUEUE:?"REDIS_JOB_QUEUE is unset"}"
: "${VIDEO_DOWNLOAD_PREFIX:?"VIDEO_DOWNLOAD_PREFIX is unset"}"
: "${VIDEO_DETAILS_PREFIX:?"VIDEO_DETAILS_PREFIX is unset"}"
: "${VIDEO_UPLOAD_URL:?"VIDEO_UPLOAD_URL is unset"}"

# loop forever
# pull job from the queue
# executing appropriate job type

while true; do
    sleep 5
    echo "getting a new job from queue"
    
    if ! job="$(get-job-from-redis)"; then
        echo "failed to get job from redis!"
        exit 1
    fi

    jobType="$(echo "${job}" | jq -r .jobType)"
    jobID="$(echo "${job}" | jq -r .jobID)"

    printf "running job (id: \"%s\") of type %s\n" "${jobID}" "${jobType}"

    echo "DEBUG: job details:"
    echo "${job}"

    if [ "${jobType}" = "transcode" ]; then
        run-transcoding-job "${job}" || echo "transcoding job failed!"
    elif [ "${jobType}" = "downscale" ]; then
        run-downscaling-job "${job}" || echo "downscaling job failed!"
    else
        echo "unknown job type: ${jobType}, ignoring it!"
    fi
done