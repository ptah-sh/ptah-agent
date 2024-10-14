#!/bin/sh

set -e

echo 'https://ptah.sh' > /tmp/check-access.txt

echo "Starting s3 upload script validation"

# Function to check if a variable is set and not empty
check_var() {
    eval value=\$$1
    if [ -z "$value" ]; then
        echo "Error: $1 is not set or is empty"
        exit 1
    fi
}

check_var ARCHIVE_FORMAT
check_var SRC_FILE_PATH
check_var DEST_FILE_PATH
check_var S3_ACCESS_KEY
check_var S3_SECRET_KEY
check_var S3_ENDPOINT
check_var S3_REGION
check_var S3_BUCKET
check_var PATH_PREFIX

SRC_FILE_PATH="/$SRC_FILE_PATH"

if [ -d "$SRC_FILE_PATH" ]; then
    cd "$SRC_FILE_PATH"
else
    cd "$(dirname "$SRC_FILE_PATH")"
fi

echo "Archiving $SRC_FILE_PATH"

SRC_FILE_PATH="/tmp/archive.$ARCHIVE_FORMAT"

case "$ARCHIVE_FORMAT" in
    "tar.gz")
        tar -czvf "$SRC_FILE_PATH" "."
        ;;
    "zip")
        apk add zip
        zip -r "$SRC_FILE_PATH" "."
        ;;
    *)
        echo "Unsupported archive format: $ARCHIVE_FORMAT"
        exit 1
        ;;
esac

echo "Uploading $SRC_FILE_PATH to s3://$S3_BUCKET/$PATH_PREFIX/$DEST_FILE_PATH"

s3cmd --guess-mime-type \
    --access_key "$S3_ACCESS_KEY" \
    --secret_key "$S3_SECRET_KEY" \
    --host "$S3_ENDPOINT" \
    --host-bucket "$S3_ENDPOINT" \
    --region "$S3_REGION" \
    put "$SRC_FILE_PATH" "s3://$S3_BUCKET/$PATH_PREFIX/$DEST_FILE_PATH"

echo "Removing $SRC_FILE_PATH/*"
rm -rf "$SRC_FILE_PATH/*" 2> /dev/null || true

echo "Done"
