#!/bin/sh
set -eu

curl --fail-with-body --request POST http://localhost:8080/incidents \
  --header 'Content-Type: application/json' \
  --data '{"work_order_id":"WO-42","area":"photo_processing","photo_id":"IMG-9","message":"thumbnail generation failed","exception":"decode: invalid image data"}'
