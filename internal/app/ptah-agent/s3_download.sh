#!/bin/sh

set -e

echo "Starting s3 download script validation"

# Function to check if a variable is set and not empty
check_var() {
    eval value=\$$1
    if [ -z "$value" ]; then
        echo "Error: $1 is not set or is empty"
        exit 1
    fi
}

check_var SRC_FILE_PATH
check_var DEST_FILE_PATH
check_var S3_ACCESS_KEY
check_var S3_SECRET_KEY
check_var S3_ENDPOINT
check_var S3_REGION
check_var S3_BUCKET
check_var PATH_PREFIX

DEST_FILE_PATH="/$DEST_FILE_PATH"

echo "Removing $DEST_FILE_PATH/*"
rm -rf "$DEST_FILE_PATH/*" 2> /dev/null || true

echo "Downloading from s3://$S3_BUCKET/$PATH_PREFIX/$SRC_FILE_PATH"

case "$SRC_FILE_PATH" in
    *".tar.gz")
        ARCHIVE_FORMAT="tar.gz"
        ;;
    *".zip")
        ARCHIVE_FORMAT="zip"
        ;;
    *)
        echo "Unsupported archive format: $SRC_FILE_PATH"
        exit 1
        ;;
esac

ARCHIVE_PATH="/tmp/archive.$ARCHIVE_FORMAT"

s3cmd --access_key "$S3_ACCESS_KEY" \
    --secret_key "$S3_SECRET_KEY" \
    --host "$S3_ENDPOINT" \
    --host-bucket "$S3_ENDPOINT" \
    --region "$S3_REGION" \
    get "s3://$S3_BUCKET/$PATH_PREFIX/$SRC_FILE_PATH" "$ARCHIVE_PATH"

echo "Extracting $SRC_FILE_PATH to $DEST_FILE_PATH"

case "$ARCHIVE_FORMAT" in
    "tar.gz")
        tar -xzvf "$ARCHIVE_PATH" -C "$DEST_FILE_PATH"
        ;;
    "zip")
        apk add zip
        unzip "$ARCHIVE_PATH"
        ;;
    *)
        echo "Unsupported archive format: $ARCHIVE_FORMAT"
        exit 1
        ;;
esac

echo "Removing $ARCHIVE_PATH"
rm -f "$ARCHIVE_PATH"

echo "Done"
