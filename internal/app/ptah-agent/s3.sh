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

check_var ARCHIVE_ENABLED
check_var SRC_FILE_PATH
check_var DEST_FILE_PATH
check_var S3_ACCESS_KEY
check_var S3_SECRET_KEY
check_var S3_ENDPOINT
check_var S3_REGION
check_var S3_BUCKET
check_var PATH_PREFIX
check_var REMOVE_SRC_FILE

if [ "$ARCHIVE_ENABLED" = "true" ]; then
    check_var ARCHIVE_FORMAT
fi

SRC_IS_DIR="false"
if [ -d "$SRC_FILE_PATH" ]; then
    SRC_IS_DIR="true"
fi

if [ "$SRC_IS_DIR" = "true" ]; then
    cd "$SRC_FILE_PATH"

    ARCHIVE_SOURCE="."
else
    cd "$(dirname "$SRC_FILE_PATH")"

    ARCHIVE_SOURCE="$(basename "$SRC_FILE_PATH")"
fi

if [ "$ARCHIVE_ENABLED" = "true" ]; then
    echo "Archiving $SRC_FILE_PATH"

    ARCHIVE_FILE_PATH="/tmp/archive.$ARCHIVE_FORMAT"

    case "$ARCHIVE_FORMAT" in
        "tar.gz")
            tar -czvf "$ARCHIVE_FILE_PATH" "$ARCHIVE_SOURCE"
            ;;
        "zip")
            apk add zip
            zip -r "$ARCHIVE_FILE_PATH" "$ARCHIVE_SOURCE"
            ;;
        *)
            echo "Unsupported archive format: $ARCHIVE_FORMAT"
            exit 1
            ;;
    esac

    # Update SRC_FILE_PATH to point to the newly created archive
    SRC_FILE_PATH="$ARCHIVE_FILE_PATH"
    SRC_IS_DIR="false"  # Archive is always a file
fi

S3_CMD_ARGS=""
if [ "$SRC_IS_DIR" = "true" ]; then
    S3_CMD_ARGS="--recursive"
    # Ensure DEST_FILE_PATH has a trailing slash
    DEST_FILE_PATH="${DEST_FILE_PATH%/}/"
else
    # Remove trailing slash if present for files
    DEST_FILE_PATH="${DEST_FILE_PATH%/}"
fi

echo "Uploading $SRC_FILE_PATH to s3://$S3_BUCKET/$PATH_PREFIX/$DEST_FILE_PATH"

SRC_UPLOAD_PATH="$SRC_FILE_PATH"
if [ "$SRC_IS_DIR" = "true" ]; then
    SRC_UPLOAD_PATH="./"
fi

s3cmd --guess-mime-type \
    --access_key "$S3_ACCESS_KEY" \
    --secret_key "$S3_SECRET_KEY" \
    --host "$S3_ENDPOINT" \
    --host-bucket "$S3_ENDPOINT" \
    --region "$S3_REGION" \
    $S3_CMD_ARGS \
    put "$SRC_UPLOAD_PATH" "s3://$S3_BUCKET/$PATH_PREFIX/$DEST_FILE_PATH"

echo "Removing $SRC_FILE_PATH/*"
if [ "$REMOVE_SRC_FILE" = "true" ]; then
    rm -rf "$SRC_FILE_PATH/*" 2> /dev/null || true
fi
echo "Done"
